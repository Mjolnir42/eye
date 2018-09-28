/*-
 * Copyright (c) 2016, Jörg Pernfuß <code.jpe@gmail.com>
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package eye // import "github.com/solnx/eye/internal/eye"

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/lib/pq"
	msg "github.com/solnx/eye/internal/eye.msg"
	"github.com/solnx/eye/lib/eye.proto/v2"
)

// LookupRead handles read requests for hash lookups
type LookupRead struct {
	Input          chan msg.Request
	Shutdown       chan struct{}
	conn           *sql.DB
	stmtCfgLookup  *sql.Stmt
	stmtActivation *sql.Stmt
	stmtPending    *sql.Stmt
	appLog         *logrus.Logger
	reqLog         *logrus.Logger
	errLog         *logrus.Logger
}

// newLookupRead return a new LookupRead handler with input buffer of length
func newLookupRead(length int) (r *LookupRead) {
	r = &LookupRead{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return
}

// process is the request dispatcher called by Run
func (r *LookupRead) process(q *msg.Request) {
	result := msg.FromRequest(q)

	switch q.Action {
	case msg.ActionConfiguration:
		r.configuration(q, &result)
	case msg.ActionActivation:
		r.activation(q, &result)
	case msg.ActionPending:
		r.pending(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// configuration returns all configurations matching a specific LookupHash
func (r *LookupRead) configuration(q *msg.Request, mr *msg.Result) {
	var (
		configurationID, dataID, configuration string
		validFrom, validUntil                  time.Time
		provisionedAt, deprovisionedAt         time.Time
		activatedAt                            pq.NullTime
		tasks                                  []string
		rows                                   *sql.Rows
		err                                    error
	)

	if rows, err = r.stmtCfgLookup.Query(
		q.LookupHash,
	); err != nil {
		mr.ServerError(err)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&configurationID,
			&dataID,
			&validFrom,
			&validUntil,
			&configuration,
			&provisionedAt,
			&deprovisionedAt,
			pq.Array(&tasks),
			&activatedAt,
		); err != nil {
			mr.ServerError(err)
			rows.Close()
			return
		}

		c := v2.Configuration{}
		if err = json.Unmarshal([]byte(configuration), &c); err != nil {
			mr.ServerError(err)
			rows.Close()
			return
		}

		if activatedAt.Valid {
			c.ActivatedAt = activatedAt.Time.Format(RFC3339Milli)
		} else {
			c.ActivatedAt = `never`
		}
		c.LookupID = q.LookupHash

		d := c.Data[0]
		d.ID = dataID
		d.Info = v2.MetaInformation{
			ValidFrom:     validFrom.Format(RFC3339Milli),
			ProvisionedAt: provisionedAt.Format(RFC3339Milli),
		}
		if msg.PosTimeInf.Equal(validUntil) {
			d.Info.ValidUntil = `infinity`
		} else {
			d.Info.ValidUntil = validUntil.Format(RFC3339Milli)
		}
		if msg.PosTimeInf.Equal(deprovisionedAt) {
			d.Info.DeprovisionedAt = `never`
		} else {
			d.Info.DeprovisionedAt = deprovisionedAt.Format(RFC3339Milli)
		}
		d.Info.Tasks = append(d.Info.Tasks, tasks...)
		tasks = []string{}
		c.Data = []v2.Data{d}
		mr.Configuration = append(mr.Configuration, c)
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err)
		return
	}
	if len(mr.Configuration) == 0 {
		mr.NotFound(fmt.Errorf(
			"Lookup for hash %s matched no configurations",
			q.LookupHash,
		))
		return
	}
	mr.OK()
}

// activation returns all configurations that have been activated after
// q.Search.Since
func (r *LookupRead) activation(q *msg.Request, mr *msg.Result) {
	var (
		rows                                          *sql.Rows
		err                                           error
		configurationID, lookupID, dataID, confResult string
		activatedAt, validFrom, validUntil            time.Time
	)

	if rows, err = r.stmtActivation.Query(
		q.Search.Since.UTC().Format(time.RFC3339Nano),
	); err != nil {
		mr.ServerError(err)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&configurationID,
			&activatedAt,
			&lookupID,
			&dataID,
			&validFrom,
			&validUntil,
			&confResult,
		); err != nil {
			rows.Close()
			mr.ServerError(err)
			return
		}
		configuration := v2.Configuration{}
		data := v2.Data{}
		if err = json.Unmarshal([]byte(confResult), &configuration); err != nil {
			rows.Close()
			mr.ServerError(err)
			return
		}
		configuration.ActivatedAt = activatedAt.Format(RFC3339Milli)
		configuration.LookupID = lookupID
		data = configuration.Data[0]
		data.ID = dataID
		data.Info = v2.MetaInformation{
			ValidFrom:  v2.FormatValidity(validFrom),
			ValidUntil: v2.FormatValidity(validUntil),
		}
		configuration.Data = []v2.Data{data}

		mr.Configuration = append(mr.Configuration, configuration)
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// pending returns all configurations that have been provisioned but not
// yet activated
func (r *LookupRead) pending(q *msg.Request, mr *msg.Result) {
	var (
		rows                        *sql.Rows
		err                         error
		configurationID, confResult string
		provisionedAt               time.Time
	)

	if rows, err = r.stmtPending.Query(
		q.Search.Since.UTC().Format(time.RFC3339Nano),
	); err != nil {
		mr.ServerError(err)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&configurationID,
			&provisionedAt,
			&confResult,
		); err != nil {
			rows.Close()
			mr.ServerError(err)
			return
		}
		configuration := v2.Configuration{}
		data := v2.Data{}
		if err = json.Unmarshal([]byte(confResult), &configuration); err != nil {
			rows.Close()
			mr.ServerError(err)
			return
		}
		configuration.ActivatedAt = `never`
		data = configuration.Data[0]
		data.Info = v2.MetaInformation{
			ProvisionedAt: v2.FormatProvision(provisionedAt),
		}
		configuration.Data = []v2.Data{data}

		mr.Configuration = append(mr.Configuration, configuration)
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
