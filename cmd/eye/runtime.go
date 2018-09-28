/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package main // import "github.com/solnx/eye/cmd/eye"

import (
	"database/sql"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/mjolnir42/erebos"
	"github.com/solnx/eye/internal/eye"
)

type runtime struct {
	conf        *erebos.Config
	conn        *sql.DB
	appLog      *logrus.Logger
	errLog      *logrus.Logger
	reqLog      *logrus.Logger
	auditLog    *logrus.Logger
	dbConnected bool
	logFileMap  *eye.LogHandleMap
}

func (run *runtime) logrotate(sigChan chan os.Signal) {
	for {
		select {
		case <-sigChan:
			for name := range run.logFileMap.Range() {
				lfHandle := run.logFileMap.Get(name)
				if err := lfHandle.Reopen(); err != nil {
					run.errLog.Errorf("Error rotating logfile %s: %s\n", name, err)
					continue
				}
				run.appLog.Printf("Rotated logfile: %s", name)
			}
		}
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
