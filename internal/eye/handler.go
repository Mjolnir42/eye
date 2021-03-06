/*-
 * Copyright (c) 2017, Jörg Pernfuß
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
)

// Handler process a specific request type
type Handler interface {
	Register(*sql.DB, ...*logrus.Logger)
	Run()
	ShutdownNow()
	Intake() chan msg.Request
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
