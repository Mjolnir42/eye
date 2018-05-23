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
	"fmt"
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
	stmtCfgHistory     *sql.Stmt
	stmtProvInfo       *sql.Stmt
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
	case msg.ActionHistory:
		r.history(q, &result)
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

// history returns the full data history for a configuration
func (r *ConfigurationRead) history(q *msg.Request, mr *msg.Result) {
	var (
		err                                     error
		tx                                      *sql.Tx
		rows                                    *sql.Rows
		validFrom, validUntil                   time.Time
		provisionTS, deprovisionTS, activatedAt time.Time
		dataID, confResult                      string
		tasks                                   []string
		configuration                           v2.Configuration
	)

	configuration.ID = q.Configuration.ID
	configuration.Data = []v2.Data{}

	// open transaction
	if tx, err = r.conn.Begin(); err != nil {
		mr.ServerError(err)
		return
	}

	// mark transaction read-only
	if _, err = tx.Exec(stmt.ReadOnlyTransaction); err != nil {
		goto abort
	}

	// query if this configurationID is activated
	if err = tx.Stmt(r.stmtActivationGet).QueryRow(
		q.Configuration.ID,
	).Scan(
		&activatedAt,
	); err == sql.ErrNoRows {
		configuration.ActivatedAt = `never`
	} else if err != nil {
		goto abort
	} else {
		configuration.ActivatedAt = activatedAt.Format(RFC3339Milli)
	}

	// read history data
	if rows, err = tx.Stmt(r.stmtCfgHistory).Query(
		q.Configuration.ID,
	); err != nil {
		goto abort
	}

	for rows.Next() {
		if err = rows.Scan(
			&dataID,
			&validFrom,
			&validUntil,
			&confResult,
		); err != nil {
			rows.Close()
			goto abort
		}

		// unmarshal JSON stored within the database
		cfg := v2.Configuration{}
		if err = json.Unmarshal([]byte(confResult), &cfg); err != nil {
			rows.Close()
			goto abort
		}

		// set configuration fields on first read
		if configuration.HostID == 0 {
			configuration.HostID = cfg.HostID
		}

		if configuration.LookupID == `` {
			configuration.LookupID = cfg.LookupID
		}

		if configuration.Metric == `` {
			configuration.Metric = cfg.Metric
		}

		// check for corrupted database on consecutive reads
		switch {
		case configuration.ID != cfg.ID:
			fallthrough
		case configuration.HostID != cfg.HostID:
			fallthrough
		case configuration.LookupID != cfg.LookupID:
			fallthrough
		case configuration.Metric != cfg.Metric:
			err = fmt.Errorf(`Data history contains corrupt modifications of immutable attributes`)
			goto abort
		}

		// populate currently retrieved metadata
		data := cfg.Data[0]
		data.Info = v2.MetaInformation{
			ValidFrom:  v2.FormatValidity(validFrom),
			ValidUntil: v2.FormatValidity(validUntil),
		}
		configuration.Data = append(configuration.Data, data)
	}
	if err = rows.Err(); err != nil {
		goto abort
	}

	// the configurationID might not exist
	if len(configuration.Data) == 0 {
		mr.NotFound(sql.ErrNoRows)
		goto rollback
	}

	// fetch provisioning information
	for idx := range configuration.Data {
		data := configuration.Data[idx]

		if err = tx.Stmt(r.stmtProvInfo).QueryRow(
			data.ID,
		).Scan(
			&provisionTS,
			&deprovisionTS,
			pq.Array(&tasks),
		); err != nil {
			// no special case for sql.ErrNoRows since no provisioning
			// record for a dataID is in fact fatal
			goto abort
		}

		data.Info.ProvisionedAt = v2.FormatProvision(provisionTS)
		data.Info.DeprovisionedAt = v2.FormatProvision(deprovisionTS)
		data.Info.Tasks = make([]string, len(tasks))
		for i := range tasks {
			data.Info.Tasks[i] = tasks[i]
		}

		configuration.Data[idx] = data
	}
	mr.Configuration = append(mr.Configuration, configuration)

	if err = tx.Commit(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
	return

abort:
	mr.ServerError(err)

rollback:
	tx.Rollback()
	return
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
