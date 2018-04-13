/*-
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
	"github.com/mjolnir42/erebos"
)

// Eye application struct
type Eye struct {
	handlerMap   *HandlerMap
	logMap       *LogHandleMap
	dbConnection *sql.DB
	conf         *erebos.Config
	appLog       *logrus.Logger
	reqLog       *logrus.Logger
	errLog       *logrus.Logger
	auditLog     *logrus.Logger
}

// New returns a new SOMA application
func New(
	appHandlerMap *HandlerMap,
	logHandleMap *LogHandleMap,
	dbConnection *sql.DB,
	conf *erebos.Config,
	appLog, reqLog, errLog, auditLog *logrus.Logger,
) *Eye {
	e := Eye{}
	e.handlerMap = appHandlerMap
	e.logMap = logHandleMap
	e.dbConnection = dbConnection
	e.conf = conf
	e.appLog = appLog
	e.reqLog = reqLog
	e.errLog = errLog
	e.auditLog = auditLog
	return &e
}

// exportLogger returns references to the instances loggers
func (e *Eye) exportLogger() []*logrus.Logger {
	return []*logrus.Logger{e.appLog, e.reqLog, e.errLog}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
