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
	"fmt"

	"github.com/Sirupsen/logrus"
	msg "github.com/solnx/eye/internal/eye.msg"
)

// DeploymentWrite handles deployment requests
type DeploymentWrite struct {
	Input      chan msg.Request
	Shutdown   chan struct{}
	conn       *sql.DB
	stmtExists *sql.Stmt
	appLog     *logrus.Logger
	reqLog     *logrus.Logger
	errLog     *logrus.Logger
}

// newDeploymentWrite return a new DeploymentWrite handler with input buffer of length
func newDeploymentWrite(length int, appLog, reqLog, errLog *logrus.Logger) (r *DeploymentWrite) {
	r = &DeploymentWrite{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	r.appLog = appLog
	r.reqLog = reqLog
	r.errLog = errLog
	return
}

// process is the request dispatcher called by Run
func (w *DeploymentWrite) process(q *msg.Request) {
	result := msg.FromRequest(q)

	// this function only sends the result in error cases. otherwise
	// w.notification() forwards the request to the configuration
	// handler which will send the result
	switch q.Action {
	case msg.ActionProcess:
		w.notification(q, &result)
	case msg.ActionNotification:
		w.notification(q, &result)
	default:
		result.UnknownRequest(q)
		q.Reply <- result
	}
}

//
func (w *DeploymentWrite) notification(q *msg.Request, mr *msg.Result) {
	var err error
	var configurationID string

	if err = w.stmtExists.QueryRow(
		q.Configuration.ID,
	).Scan(
		&configurationID,
	); err != nil && err != sql.ErrNoRows {
		w.appLog.Errorln(err)
		mr.ServerError(err)
		q.Reply <- *mr
		return
	}

	// get the configuration update handler
	handler := handlerLookup.Get(`configuration_w`)

	// check if we have the configuration
	switch q.ConfigurationTask {
	case msg.TaskRollout:
		// rollout + configuration does not exist -> ConfigurationAdd
		if err == sql.ErrNoRows {
			q.Section = msg.SectionConfiguration
			q.Action = msg.ActionAdd
			handler.Intake() <- *q
			return
		} else if err != nil {
			mr.ServerError(err)
			q.Reply <- *mr
			return
		}
	case msg.TaskDelete, msg.TaskDeprovision:
		// deprovision|delete + configuration does not exist -> no-op
		if err == sql.ErrNoRows {
			q.Section = msg.SectionConfiguration
			q.Action = msg.ActionNop
			handler.Intake() <- *q
			return
		} else if err != nil {
			mr.ServerError(err)
			q.Reply <- *mr
			return
		}
	}

	if q.Configuration.ID != configurationID {
		panic(fmt.Sprintf(
			"Database corrupt! Lookup for %s found %s",
			q.Configuration.ID,
			configurationID,
		))
	}

	q.Section = msg.SectionConfiguration
	switch q.ConfigurationTask {
	case msg.TaskRollout:
		q.Action = msg.ActionUpdate
	case msg.TaskDelete, msg.TaskDeprovision:
		q.Action = msg.ActionRemove
	}
	handler.Intake() <- *q
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
