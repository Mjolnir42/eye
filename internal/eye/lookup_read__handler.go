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

	"github.com/Sirupsen/logrus"
	msg "github.com/solnx/eye/internal/eye.msg"
	stmt "github.com/solnx/eye/internal/eye.stmt"
)

// Implementation of the Handler interface

// Register initializes resources provided by the eye application
func (r *LookupRead) Register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
}

// Run is the event loop for LookupRead
func (r *LookupRead) Run() {
	var err error

	for statement, prepStmt := range map[string]**sql.Stmt{
		stmt.LookupActivation:    &r.stmtActivation,
		stmt.LookupConfiguration: &r.stmtCfgLookup,
		stmt.LookupPending:       &r.stmtPending,
	} {
		if *prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.appLog.Fatal(`lookup`, err, stmt.Name(statement))
		}
		defer (*prepStmt).Close()
	}

runloop:
	for {
		select {
		case <-r.Shutdown:
			break runloop
		case req := <-r.Input:
			go func() {
				r.process(&req)
			}()
		}
	}
}

// ShutdownNow signals the handler to shut down
func (r *LookupRead) ShutdownNow() {
	close(r.Shutdown)
}

// Intake exposes the Input channel as part of the handler interface
func (r *LookupRead) Intake() chan msg.Request {
	return r.Input
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
