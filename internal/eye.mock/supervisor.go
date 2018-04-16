/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package mock // import "github.com/mjolnir42/eye/internal/eye.mock"

import (
	"database/sql"

	"github.com/Sirupsen/logrus"
	"github.com/mjolnir42/erebos"
	msg "github.com/mjolnir42/eye/internal/eye.msg"
)

// PermissiveSupervisor is a special supervisor that permits all
// requests
type PermissiveSupervisor struct {
	Input    chan msg.Request
	Update   chan msg.Request
	Shutdown chan struct{}
	conn     *sql.DB
	appLog   *logrus.Logger
	reqLog   *logrus.Logger
	errLog   *logrus.Logger
	auditLog *logrus.Logger
	conf     *erebos.Config
}

// NewPermissiveSupervisor returns a new PermissiveSupervisor
func NewPermissiveSupervisor(c *erebos.Config) *PermissiveSupervisor {
	s := &PermissiveSupervisor{}
	s.conf = c
	s.Input = make(chan msg.Request, s.conf.Eye.QueueLen)
	s.Update = make(chan msg.Request, s.conf.Eye.QueueLen)
	s.Shutdown = make(chan struct{})
	return s
}

// Register initializes resources provided by the Eye app
func (s *PermissiveSupervisor) Register(c *sql.DB, l ...*logrus.Logger) {
	s.conn = c
	s.appLog = l[0]
	s.reqLog = l[1]
	s.errLog = l[2]
}

// RegisterAuditLog initializes the audit log provided by the Soma app
func (s *PermissiveSupervisor) RegisterAuditLog(a *logrus.Logger) {
	s.auditLog = a
}

// Intake exposes the Input channel as part of the handler interface
func (s *PermissiveSupervisor) Intake() chan msg.Request {
	return s.Input
}

// Run is the event loop for PermissiveSupervisor
func (s *PermissiveSupervisor) Run() {

runloop:
	for {
		select {
		case <-s.Update:
			// ignore permission cache updates
		case <-s.Shutdown:
			break runloop
		case req := <-s.Input:
			s.process(&req)
		}
	}
}

// process is the event dispatcher
func (s *PermissiveSupervisor) process(q *msg.Request) {
	switch q.Section {
	case msg.SectionSupervisor:
		switch q.Action {
		case msg.ActionAuthenticate:
			go func() { s.authenticate(q) }()
		case msg.ActionAuthorize:
			go func() { s.authorize(q) }()
		}
	}
}

// ShutdownNow signals the handler to stop
func (s *PermissiveSupervisor) ShutdownNow() {
	close(s.Shutdown)
}

// authenticate handles supervisor requests for authentication
func (s *PermissiveSupervisor) authenticate(q *msg.Request) {
	result := msg.FromRequest(q)
	result.Code = msg.ResultUnauthorized
	result.Super.Verdict = msg.VerdictUnauthorized

	switch q.Super.Task {
	case msg.TaskBasicAuth:
		s.authenticateBasicAuth(q, &result)
	}

	q.Reply <- result
}

// authenticateBasicAuth performs BasicAuth authentication
func (s *PermissiveSupervisor) authenticateBasicAuth(q *msg.Request, mr *msg.Result) {
	mr.Super.Verdict = msg.VerdictOK
	mr.OK()
}

// authorize handles supervisor requests for authorization
func (s *PermissiveSupervisor) authorize(q *msg.Request) {
	result := msg.FromRequest(q)
	result.Super.Verdict = msg.VerdictOK
	result.OK()

	q.Reply <- result
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
