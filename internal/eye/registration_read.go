/*-
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
	"github.com/solnx/eye/lib/eye.proto/v2"
)

// RegistrationRead handles read requests for hash lookups
type RegistrationRead struct {
	Input      chan msg.Request
	Shutdown   chan struct{}
	conn       *sql.DB
	stmtList   *sql.Stmt
	stmtSearch *sql.Stmt
	stmtShow   *sql.Stmt
	appLog     *logrus.Logger
	reqLog     *logrus.Logger
	errLog     *logrus.Logger
}

// newRegistrationRead return a new RegistrationRead handler with input buffer of length
func newRegistrationRead(length int) (r *RegistrationRead) {
	r = &RegistrationRead{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return
}

// process is the request dispatcher called by Run
func (r *RegistrationRead) process(q *msg.Request) {
	result := msg.FromRequest(q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	case msg.ActionSearch, msg.ActionRegistration:
		r.search(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// list returns all registrations by ID
func (r *RegistrationRead) list(q *msg.Request, mr *msg.Result) {
	var (
		registrationID string
		rows           *sql.Rows
		err            error
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&registrationID); err != nil {
			rows.Close()
			mr.ServerError(err)
			return
		}
		mr.Registration = append(mr.Registration, v2.Registration{
			ID: registrationID,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// show returns a specific registration
func (r *RegistrationRead) show(q *msg.Request, mr *msg.Result) {
	var (
		err                                  error
		registrationID, application, address string
		port, database                       int64
		registeredAt                         time.Time
	)

	if err = r.stmtShow.QueryRow(
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
		return
	} else if err != nil {
		mr.ServerError(err)
		return
	}
	mr.Registration = append(mr.Registration, v2.Registration{
		ID:           registrationID,
		Application:  application,
		Address:      address,
		Port:         port,
		Database:     database,
		RegisteredAt: registeredAt,
	})
	mr.OK()
}

// search returns all registrations that fulfull a specific search
// condition
func (r *RegistrationRead) search(q *msg.Request, mr *msg.Result) {
	var (
		rows                                 *sql.Rows
		err                                  error
		searchApp, searchAddr                sql.NullString
		searchPort, searchDB                 sql.NullInt64
		registrationID, application, address string
		port, database                       int64
		registeredAt                         time.Time
	)

	// set NULL-able query conditions
	if q.Search.Registration.Application != `` {
		searchApp.String = q.Search.Registration.Application
		searchApp.Valid = true
	}
	if q.Search.Registration.Address != `` {
		searchAddr.String = q.Search.Registration.Address
		searchAddr.Valid = true
	}
	if q.Search.Registration.Port != 0 {
		searchPort.Int64 = q.Search.Registration.Port
		searchPort.Valid = true
	}
	if q.Search.Registration.Database != -1 {
		// REST encodes unspecified database parameter as -1 since 0 is
		// valid
		searchDB.Int64 = q.Search.Registration.Database
		searchDB.Valid = true
	}

	// perform search
	if rows, err = r.stmtSearch.Query(
		searchApp,
		searchAddr,
		searchPort,
		searchDB,
	); err != nil {
		mr.ServerError(err)
		return
	}

	// iterate over result list
	for rows.Next() {
		if err = rows.Scan(
			&registrationID,
			&application,
			&address,
			&port,
			&database,
			&registeredAt,
		); err != nil {
			rows.Close()
			mr.ServerError(err)
			return
		}
		// build result list
		mr.Registration = append(mr.Registration, v2.Registration{
			ID:           registrationID,
			Application:  application,
			Address:      address,
			Port:         port,
			Database:     database,
			RegisteredAt: registeredAt,
		})
	}

	// check for errors which occurred during iteration
	if err = rows.Err(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
