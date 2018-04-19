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
	proto "github.com/mjolnir42/eye/lib/eye.proto"
)

// ConfigurationWrite handles read requests for hash lookups
type ConfigurationWrite struct {
	Input                             chan msg.Request
	Shutdown                          chan struct{}
	conn                              *sql.DB
	stmtConfigurationAdd              *sql.Stmt
	stmtConfigurationCountForLookupID *sql.Stmt
	stmtConfigurationRemove           *sql.Stmt
	stmtConfigurationShow             *sql.Stmt
	stmtConfigurationUpdate           *sql.Stmt
	stmtLookupAdd                     *sql.Stmt
	stmtLookupIDForConfiguration      *sql.Stmt
	stmtLookupRemove                  *sql.Stmt
	appLog                            *logrus.Logger
	reqLog                            *logrus.Logger
	errLog                            *logrus.Logger
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
	case msg.ActionRemove:
		w.remove(q, &result)
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

// remove deletes a configuration from the database
func (w *ConfigurationWrite) remove(q *msg.Request, mr *msg.Result) {
	var (
		err                  error
		tx                   *sql.Tx
		res                  sql.Result
		lookupID, confResult string
		popcnt               int
		configuration        proto.Configuration
	)

	// open transaction
	if tx, err = w.conn.Begin(); err != nil {
		mr.ServerError(err)
		return
	}

	// retrieve lookupID for Configuration.ID prior to deleting it; if
	// this is the last configuration using this hash then the
	// lookup is deleted as well
	if err = tx.Stmt(w.stmtLookupIDForConfiguration).QueryRow(
		q.Configuration.ID,
	).Scan(
		&lookupID,
	); err == sql.ErrNoRows {
		// not being able to delete what we do not have is ok
		goto commitTx
	} else if err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}

	// retrieve full configuration prior to deleting it; this is
	// required for requests with q.ConfigurationTask set to
	// msg.TaskDelete and enabled AlarmClearing so that the OK event can
	// be constructed with the correct metadata
	if err = tx.Stmt(w.stmtConfigurationShow).QueryRow(
		q.Configuration.ID,
	).Scan(
		&confResult,
	); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}
	if err = json.Unmarshal([]byte(confResult), &configuration); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}
	mr.Configuration = append(mr.Configuration, configuration)

	// delete configuration
	if res, err = tx.Stmt(w.stmtConfigurationRemove).Exec(
		q.Configuration.ID,
	); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}
	// statement should affect 0 or 1 rows
	if count, _ := res.RowsAffected(); count > 1 || count < 0 {
		mr.ServerError(fmt.Errorf("Rollback: delete statement affected %d rows", count))
		tx.Rollback()
		return
	}

	// check number of remaining configurations using the same lookupID
	if err = tx.Stmt(w.stmtConfigurationCountForLookupID).QueryRow(
		lookupID,
	).Scan(
		&popcnt,
	); err != nil {
		// SELECT COUNT queries must always return a result row,
		// so sql.ErrNoRows is fatal
		mr.ServerError(err)
		tx.Rollback()
		return
	}
	if popcnt > 0 {
		// there are still configurations using the the lookupID, skip
		// deleting lookupID and commit transaction
		goto commitTx
	}

	// delete lookupID entry
	if res, err = tx.Stmt(w.stmtLookupRemove).Exec(
		lookupID,
	); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}
	// statement should affect 1 row
	if count, _ := res.RowsAffected(); count != 1 {
		mr.ServerError(fmt.Errorf("Rollback: delete statement affected %d rows", count))
		tx.Rollback()
		return
	}

commitTx:
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
