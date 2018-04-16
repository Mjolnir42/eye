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

// LookupRead handles read requests for hash lookups
type LookupRead struct {
	Input      chan msg.Request
	Shutdown   chan struct{}
	conn       *sql.DB
	stmtSearch *sql.Stmt
	appLog     *logrus.Logger
	reqLog     *logrus.Logger
	errLog     *logrus.Logger
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
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// configuration returns all configurations matching a specific LookupHash
func (r *LookupRead) configuration(q *msg.Request, mr *msg.Result) {
	var (
		configuration string
		rows          *sql.Rows
		err           error
	)

	if rows, err = r.stmtSearch.Query(q.LookupHash); err != nil {
		mr.ServerError(err)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&configuration); err != nil {
			mr.ServerError(err)
			return
		}

		c := proto.Configuration{}
		if err = json.Unmarshal([]byte(configuration), &c); err != nil {
			mr.ServerError(err)
			return
		}
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

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
