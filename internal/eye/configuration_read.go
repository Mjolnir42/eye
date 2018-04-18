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

	"github.com/Sirupsen/logrus"
	msg "github.com/mjolnir42/eye/internal/eye.msg"
	proto "github.com/mjolnir42/eye/lib/eye.proto"
)

// ConfigurationRead handles read requests for hash lookups
type ConfigurationRead struct {
	Input    chan msg.Request
	Shutdown chan struct{}
	conn     *sql.DB
	stmtList *sql.Stmt
	stmtShow *sql.Stmt
	appLog   *logrus.Logger
	reqLog   *logrus.Logger
	errLog   *logrus.Logger
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

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&configurationID); err != nil {
			rows.Close()
			mr.ServerError(err)
			return
		}
		mr.Configuration = append(mr.Configuration, proto.Configuration{
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
		err           error
		confResult    string
		configuration proto.Configuration
	)

	if err = r.stmtShow.QueryRow(
		q.Configuration.ID,
	).Scan(
		&confResult,
	); err == sql.ErrNoRows {
		mr.NotFound(err)
		return
	} else if err != nil {
		mr.ServerError(err)
		return
	}
	if err = json.Unmarshal([]byte(confResult), &configuration); err != nil {
		mr.ServerError(err)
		return
	}
	mr.Configuration = append(mr.Configuration, configuration)
	mr.OK()
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
