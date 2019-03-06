/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

// Package rest implements the REST routes to access EYE.
package rest // import "github.com/solnx/eye/internal/eye.rest"

import (
	"database/sql"
	"net/http"
	"net/url"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/mjolnir42/erebos"
	"github.com/mjolnir42/limit"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/solnx/eye/internal/eye"
	msg "github.com/solnx/eye/internal/eye.msg"
	stmt "github.com/solnx/eye/internal/eye.stmt"
	wall "github.com/solnx/eye/lib/eye.wall"
)

// ShutdownInProgress indicates a pending service shutdown
var ShutdownInProgress bool

// Metrics is the map of runtime metric registries
var Metrics = make(map[string]metrics.Registry)

// Rest holds the required state for the REST interface
type Rest struct {
	isAuthorized func(*msg.Request) bool
	handlerMap   *eye.HandlerMap
	conf         *erebos.Config
	restricted   bool

	conn               *sql.DB
	stmtRegisterGetAll *sql.Stmt
	// concurrenyLimit caps the number of active outgoing HTTP requests
	limit *limit.Limit
	// notification template
	tmpl *template.Template
	// cache invalidator
	invl   *wall.Invalidation
	appLog *logrus.Logger
}

// New returns a new REST interface
func New(
	authorizationFunction func(*msg.Request) bool,
	appHandlerMap *eye.HandlerMap,
	conf *erebos.Config,
	conn *sql.DB,
	appLog *logrus.Logger,
) *Rest {
	x := Rest{}
	x.isAuthorized = authorizationFunction
	x.restricted = false
	x.handlerMap = appHandlerMap
	x.conf = conf
	x.limit = limit.New(conf.Eye.ConcurrencyLimit)
	x.tmpl = template.Must(template.ParseFiles(conf.Eye.AlarmTemplateFile))
	x.invl = wall.NewInvalidation(conf)
	x.appLog = appLog
	x.conn = conn
	var err error
	for statement, prepStmt := range map[string]**sql.Stmt{
		stmt.RegistryGetAll: &x.stmtRegisterGetAll,
	} {
		if *prepStmt, err = x.conn.Prepare(statement); err != nil {
			x.appLog.Fatal(`lookup`, err, stmt.Name(statement))
		}
		defer (*prepStmt).Close()
	}
	return &x
}

// Run is the event server for Rest
func (x *Rest) Run() {
	router := x.setupRouter()
	x.conf.Eye.Daemon.URL = &url.URL{}
	x.conf.Eye.Daemon.URL.Host = x.conf.Eye.Daemon.Listen + ":" + x.conf.Eye.Daemon.Port
	// TODO switch to new abortable interface
	if x.conf.Eye.Daemon.TLS {
		// XXX log.Fatal
		http.ListenAndServeTLS(
			x.conf.Eye.Daemon.URL.Host,
			x.conf.Eye.Daemon.Cert,
			x.conf.Eye.Daemon.Key,
			router,
		)
	} else {
		// XXX log.Fatal
		http.ListenAndServe(x.conf.Eye.Daemon.URL.Host, router)
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
