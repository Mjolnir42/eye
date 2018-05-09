/*-
 * Copyright (c) 2016, Jörg Pernfuß <code.jpe@gmail.com>
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package eye // import "github.com/mjolnir42/eye/internal/eye"

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/lib/pq"
	msg "github.com/mjolnir42/eye/internal/eye.msg"
	stmt "github.com/mjolnir42/eye/internal/eye.stmt"
	"github.com/mjolnir42/eye/lib/eye.proto/v2"
)

// ConfigurationRead handles read requests for hash lookups
type ConfigurationRead struct {
	Input              chan msg.Request
	Shutdown           chan struct{}
	conn               *sql.DB
	stmtCfgSelectValid *sql.Stmt
	stmtCfgShow        *sql.Stmt
	stmtActivationGet  *sql.Stmt
	stmtCfgList        *sql.Stmt
	appLog             *logrus.Logger
	reqLog             *logrus.Logger
	errLog             *logrus.Logger
}

// newConfigurationRead return a new ConfigurationRead handler with input buffer of length
func newConfigurationRead(length int) (r *ConfigurationRead) {
	r = &ConfigurationRead{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return
}

// process is the request dispatcher called by Run
func (r *ConfigurationRead) process(q *msg.Request) {
	result := msg.FromRequest(q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// list returns all configurations by ID
func (r *ConfigurationRead) list(q *msg.Request, mr *msg.Result) {
	var (
		configurationID string
		rows            *sql.Rows
		err             error
	)

	if rows, err = r.stmtCfgList.Query(); err != nil {
		mr.ServerError(err)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&configurationID); err != nil {
			rows.Close()
			mr.ServerError(err)
			return
		}
		mr.Configuration = append(mr.Configuration, v2.Configuration{
			ID: configurationID,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// show returns a specific configuration
func (r *ConfigurationRead) show(q *msg.Request, mr *msg.Result) {
	var (
		err                                     error
		dataID, confResult                      string
		tasks                                   []string
		configuration                           v2.Configuration
		data                                    v2.Data
		validFrom, validUntil                   time.Time
		provisionTS, deprovisionTS, activatedAt time.Time
		tx                                      *sql.Tx
	)

	// open transaction
	if tx, err = r.conn.Begin(); err != nil {
		mr.ServerError(err)
		return
	}

	// mark transaction read-only
	if _, err = tx.Exec(stmt.ReadOnlyTransaction); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}

	// get currently valid dataID
	if err = tx.Stmt(r.stmtCfgSelectValid).QueryRow(
		q.Configuration.ID,
	).Scan(
		&dataID,
		&validFrom,
	); err == sql.ErrNoRows {
		mr.NotFound(err)
		tx.Rollback()
		return
	} else if err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}

	// read queried dataID
	if err = tx.Stmt(r.stmtCfgShow).QueryRow(
		dataID,
	).Scan(
		&confResult,
		&validUntil,
		&provisionTS,
		&deprovisionTS,
		pq.Array(&tasks),
	); err == sql.ErrNoRows {
		mr.NotFound(err)
		tx.Rollback()
		return
	} else if err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}

	// unmarshal JSON stored within the database
	if err = json.Unmarshal([]byte(confResult), &configuration); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}

	// query if this configurationID is activated
	if err = tx.Stmt(r.stmtActivationGet).QueryRow(
		q.Configuration.ID,
	).Scan(
		&activatedAt,
	); err == sql.ErrNoRows {
		configuration.ActivatedAt = `never`
	} else if err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	} else {
		configuration.ActivatedAt = activatedAt.Format(RFC3339Milli)
	}

	// populate result metadata
	data = configuration.Data[0]
	data.Info = v2.MetaInformation{
		ValidFrom:     validFrom.Format(RFC3339Milli),
		ProvisionedAt: provisionTS.Format(RFC3339Milli),
	}
	if msg.PosTimeInf.Equal(deprovisionTS) {
		data.Info.DeprovisionedAt = `never`
	} else {
		data.Info.DeprovisionedAt = deprovisionTS.Format(RFC3339Milli)
	}
	if msg.PosTimeInf.Equal(validUntil) {
		data.Info.ValidUntil = `infinity`
	} else {
		data.Info.ValidUntil = validUntil.Format(RFC3339Milli)
	}
	data.Info.Tasks = append(data.Info.Tasks, tasks...)
	configuration.Data = []v2.Data{data}
	mr.Configuration = append(mr.Configuration, configuration)

	if err = tx.Commit(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
