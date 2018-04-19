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

	"github.com/Sirupsen/logrus"
	msg "github.com/mjolnir42/eye/internal/eye.msg"
)

// ConfigurationWrite handles read requests for hash lookups
type ConfigurationWrite struct {
	Input                   chan msg.Request
	Shutdown                chan struct{}
	conn                    *sql.DB
	stmtConfigurationAdd    *sql.Stmt
	stmtConfigurationRemove *sql.Stmt
	stmtConfigurationUpdate *sql.Stmt
	stmtLookupAdd           *sql.Stmt
	appLog                  *logrus.Logger
	reqLog                  *logrus.Logger
	errLog                  *logrus.Logger
}

// newConfigurationWrite return a new ConfigurationWrite handler with input buffer of length
func newConfigurationWrite(length int) (w *ConfigurationWrite) {
	w = &ConfigurationWrite{}
	w.Input = make(chan msg.Request, length)
	w.Shutdown = make(chan struct{})
	return
}

// process is the request dispatcher called by Run
func (w *ConfigurationWrite) process(q *msg.Request) {
	result := msg.FromRequest(q)

	switch q.Action {
	case msg.ActionAdd:
		w.add(q, &result)
	case msg.ActionUpdate:
		w.update(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// add inserts a configuration profile into the database
func (w *ConfigurationWrite) add(q *msg.Request, mr *msg.Result) {
	var (
		err   error
		tx    *sql.Tx
		jsonb []byte
		res   sql.Result
	)

	if jsonb, err = json.Marshal(q.Configuration); err != nil {
		mr.ServerError(err)
		return
	}

	if tx, err = w.conn.Begin(); err != nil {
		mr.ServerError(err)
		return
	}

	if res, err = tx.Stmt(w.stmtLookupAdd).Exec(
		q.LookupHash,
		int(q.Configuration.HostID),
		q.Configuration.Metric,
	); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}
	// Statement should affect 1 row for the first configuration with
	// this lookupID and 0 rows afterwards for additional configurations
	if count, _ := res.RowsAffected(); count > 1 || count < 0 {
		mr.ServerError(fmt.Errorf("Insert statement affected %d rows", count))
		tx.Rollback()
		return
	}

	if res, err = tx.Stmt(w.stmtConfigurationAdd).Exec(
		q.Configuration.ID,
		q.LookupHash,
		jsonb,
	); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}
	// statement should affect 1 row
	if count, _ := res.RowsAffected(); count != 1 {
		mr.ServerError(fmt.Errorf("Rollback: insert statement affected %d rows", count))
		tx.Rollback()
		return
	}

	if err = tx.Commit(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// update replaces a configuration
func (w *ConfigurationWrite) update(q *msg.Request, mr *msg.Result) {
	var (
		err   error
		tx    *sql.Tx
		jsonb []byte
		res   sql.Result
	)

	if jsonb, err = json.Marshal(q.Configuration); err != nil {
		mr.ServerError(err)
		return
	}

	if tx, err = w.conn.Begin(); err != nil {
		mr.ServerError(err)
		return
	}

	if res, err = tx.Stmt(w.stmtConfigurationUpdate).Exec(
		q.Configuration.ID,
		q.LookupHash,
		jsonb,
	); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}
	// statement should affect 1 row
	if count, _ := res.RowsAffected(); count != 1 {
		mr.ServerError(fmt.Errorf("Rollback: update statement affected %d rows", count))
		tx.Rollback()
		return
	}

	if err = tx.Commit(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
