/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

// Package eye implements the application logic of the eye server
package eye // import "github.com/solnx/eye/internal/eye"

import (
	"database/sql"

	"github.com/Sirupsen/logrus"
	"github.com/mjolnir42/erebos"
)

// RFC3339Milli is a format string for millisecond precision RFC3339
const RFC3339Milli string = "2006-01-02T15:04:05.000Z07:00"

// handlerLookup is used by eye handlers to communicate with each other
var handlerLookup *HandlerMap

// Eye application struct
type Eye struct {
	handlerMap   *HandlerMap
	dbConnection *sql.DB
	conf         *erebos.Config
	appLog       *logrus.Logger
}

// New returns a new SOMA application
func New(
	appHandlerMap *HandlerMap,
	dbConnection *sql.DB,
	conf *erebos.Config,
	appLog *logrus.Logger,
) *Eye {
	e := Eye{}
	e.handlerMap = appHandlerMap
	e.dbConnection = dbConnection
	e.conf = conf
	e.appLog = appLog
	handlerLookup = appHandlerMap
	return &e
}

// exportLogger returns references to the instances loggers
func (e *Eye) exportLogger() []*logrus.Logger {
	return []*logrus.Logger{e.appLog}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
