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

	"github.com/Sirupsen/logrus"
	msg "github.com/mjolnir42/eye/internal/eye.msg"
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
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
