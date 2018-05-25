/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

// Package rest implements the REST routes to access EYE.
package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import (
	"net/http"

	"github.com/mjolnir42/erebos"
	"github.com/mjolnir42/eye/internal/eye"
	msg "github.com/mjolnir42/eye/internal/eye.msg"
	"github.com/mjolnir42/limit"
	metrics "github.com/rcrowley/go-metrics"
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
	// concurrenyLimit caps the number of active outgoing HTTP requests
	limit *limit.Limit
}

// New returns a new REST interface
func New(
	authorizationFunction func(*msg.Request) bool,
	appHandlerMap *eye.HandlerMap,
	conf *erebos.Config,
) *Rest {
	x := Rest{}
	x.isAuthorized = authorizationFunction
	x.restricted = false
	x.handlerMap = appHandlerMap
	x.conf = conf
	x.limit = limit.New(conf.Eye.ConcurrencyLimit)
	return &x
}

// Run is the event server for Rest
func (x *Rest) Run() {
	router := x.setupRouter()

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
