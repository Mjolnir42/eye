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
func (w *ConfigurationWrite) Register(c *sql.DB, l ...*logrus.Logger) {
	w.conn = c
	w.appLog = l[0]
	w.reqLog = l[1]
	w.errLog = l[2]
}

// Run is the event loop for ConfigurationWrite
func (w *ConfigurationWrite) Run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.LookupAddID:             w.stmtLookupAddID,
		stmt.CfgAddID:                w.stmtCfgAddID,
		stmt.CfgSelectValidForUpdate: w.stmtCfgSelectValidForUpdate,
		stmt.CfgDataUpdateValidity:   w.stmtCfgDataUpdateValidity,
		stmt.CfgAddData:              w.stmtCfgAddData,
		stmt.ProvAdd:                 w.stmtProvAdd,
		stmt.ActivationGet:           w.stmtActivationGet,
		stmt.ProvFinalize:            w.stmtProvFinalize,
		stmt.ActivationDel:           w.stmtActivationDel,
		stmt.CfgShow:                 w.stmtCfgShow,
		stmt.ActivationSet:           w.stmtActivationSet,
	} {
		if prepStmt, err = w.conn.Prepare(statement); err != nil {
			w.errLog.Fatal(`lookup`, err, stmt.Name(statement))
		}
		defer prepStmt.Close()
		switch statement {
		case stmt.LookupAddID:
			w.stmtLookupAddID = prepStmt
		case stmt.CfgAddID:
			w.stmtCfgAddID = prepStmt
		case stmt.CfgSelectValidForUpdate:
			w.stmtCfgSelectValidForUpdate = prepStmt
		case stmt.CfgDataUpdateValidity:
			w.stmtCfgDataUpdateValidity = prepStmt
		case stmt.CfgAddData:
			w.stmtCfgAddData = prepStmt
		case stmt.ProvAdd:
			w.stmtProvAdd = prepStmt
		case stmt.ActivationGet:
			w.stmtActivationGet = prepStmt
		case stmt.CfgShow:
			w.stmtCfgShow = prepStmt
		case stmt.ActivationSet:
			w.stmtActivationSet = prepStmt
		}
	}

runloop:
	for {
		select {
		case <-w.Shutdown:
			break runloop
		case req := <-w.Input:
			go func() {
				w.process(&req)
			}()
		}
	}
}

// ShutdownNow signals the handler to shut down
func (w *ConfigurationWrite) ShutdownNow() {
	close(w.Shutdown)
}

// Intake exposes the Input channel as part of the handler interface
func (w *ConfigurationWrite) Intake() chan msg.Request {
	return w.Input
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
