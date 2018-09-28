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
	"time"

	"github.com/Sirupsen/logrus"
	msg "github.com/solnx/eye/internal/eye.msg"
	uuid "github.com/satori/go.uuid"
)

// RegistrationWrite handles read requests for hash lookups
type RegistrationWrite struct {
	Input      chan msg.Request
	Shutdown   chan struct{}
	conn       *sql.DB
	stmtAdd    *sql.Stmt
	stmtRemove *sql.Stmt
	stmtShow   *sql.Stmt
	stmtUpdate *sql.Stmt
	appLog     *logrus.Logger
	reqLog     *logrus.Logger
	errLog     *logrus.Logger
}

// newRegistrationWrite return a new RegistrationWrite handler with input buffer of length
func newRegistrationWrite(length int) (w *RegistrationWrite) {
	w = &RegistrationWrite{}
	w.Input = make(chan msg.Request, length)
	w.Shutdown = make(chan struct{})
	return
}

// process is the request dispatcher called by Run
func (w *RegistrationWrite) process(q *msg.Request) {
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

// add inserts a registration into the database
func (w *RegistrationWrite) add(q *msg.Request, mr *msg.Result) {
	var res sql.Result
	var err error

	// generate RegistrationID
	if q.Registration.ID, err = func() (string, error) {
		u, e := uuid.NewV4()
		if e != nil {
			return ``, e
		}
		return u.String(), nil
	}(); err != nil {
		mr.ServerError(err)
		return
	}

	// insert registration into the database
	if res, err = w.stmtAdd.Exec(
		q.Registration.ID,
		q.Registration.Application,
		q.Registration.Address,
		q.Registration.Port,
		q.Registration.Database,
	); err != nil {
		mr.ServerError(err)
		return
	}

	// set OK based on affected rows
	if mr.RowCnt(res.RowsAffected()) {
		mr.Registration = append(mr.Registration, q.Registration)
	}
}

// remove deletes a registration from the database
func (w *RegistrationWrite) remove(q *msg.Request, mr *msg.Result) {
	var (
		tx                                   *sql.Tx
		res                                  sql.Result
		err                                  error
		registrationID, application, address string
		port, database                       int64
		registeredAt                         time.Time
	)

	// open transaction
	if tx, err = w.conn.Begin(); err != nil {
		mr.ServerError(err)
		return
	}

	// retrieve full configuration so we can return what has been deleted
	if err = tx.Stmt(w.stmtShow).QueryRow(
		q.Registration.ID,
	).Scan(
		&registrationID,
		&application,
		&address,
		&port,
		&database,
		&registeredAt,
	); err == sql.ErrNoRows {
		mr.NotFound(err)
		tx.Rollback()
		return
	} else if err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}

	if q.Registration.ID != registrationID {
		mr.ServerError(nil)
		tx.Rollback()
		return
	}
	q.Registration.Application = application
	q.Registration.Address = address
	q.Registration.Port = port
	q.Registration.Database = database
	q.Registration.RegisteredAt = registeredAt

	// delete registration
	if res, err = tx.Stmt(w.stmtRemove).Exec(
		q.Registration.ID,
	); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}

	// check result and close transaction
	if mr.ExpectedRows(&res, 1) {
		if err = tx.Commit(); err != nil {
			mr.ServerError(err)
			return
		}
		mr.Registration = append(mr.Registration, q.Registration)
		return
	}
	tx.Rollback()
}

// update replaces a registration
func (w *RegistrationWrite) update(q *msg.Request, mr *msg.Result) {
	var (
		tx  *sql.Tx
		res sql.Result
		err error
	)

	// open transaction
	if tx, err = w.conn.Begin(); err != nil {
		mr.ServerError(err)
		return
	}

	// update registration
	q.Registration.RegisteredAt = time.Now().UTC()
	if res, err = tx.Stmt(w.stmtUpdate).Exec(
		q.Registration.ID,
		q.Registration.Application,
		q.Registration.Address,
		q.Registration.Port,
		q.Registration.Database,
		q.Registration.RegisteredAt,
	); err != nil {
		mr.ServerError(err)
		tx.Rollback()
		return
	}

	// check result and close transaction
	if mr.ExpectedRows(&res, 1) {
		if err = tx.Commit(); err != nil {
			mr.ServerError(err)
			return
		}
		mr.Registration = append(mr.Registration, q.Registration)
		return
	}
	tx.Rollback()
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
