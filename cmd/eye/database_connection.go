/*-
 * Copyright (c) 2016,2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package main // import "github.com/solnx/eye/cmd/eye"

import (
	"database/sql"
	"fmt"

	"time"

	"github.com/lib/pq"
	msg "github.com/solnx/eye/internal/eye.msg"
	stmt "github.com/solnx/eye/internal/eye.stmt"
)

// connectDatabase opens the connection to the database and configures
// the established connection
func (run *runtime) connectDatabase() {
	var err error
	var rows *sql.Rows
	var schema string
	var version int64

	driver := `postgres`

	run.appLog.Printf("Using postgreSQL: dbname='%s' user='%s' host='%s' port='%s' sslmode='%s' connect_timeout='%s' options='-c timezone=UTC'",
		run.conf.PostgreSQL.Name,
		run.conf.PostgreSQL.User,
		run.conf.PostgreSQL.Host,
		run.conf.PostgreSQL.Port,
		run.conf.PostgreSQL.TLSMode,
		run.conf.PostgreSQL.Timeout,
	)
	connect := fmt.Sprintf("dbname='%s' user='%s' password='%s' host='%s' port='%s' sslmode='%s' connect_timeout='%s' options='-c timezone=UTC'",
		run.conf.PostgreSQL.Name,
		run.conf.PostgreSQL.User,
		run.conf.PostgreSQL.Pass,
		run.conf.PostgreSQL.Host,
		run.conf.PostgreSQL.Port,
		run.conf.PostgreSQL.TLSMode,
		run.conf.PostgreSQL.Timeout,
	)

	// enable handling of infinity timestamps
	pq.EnableInfinityTs(msg.NegTimeInf, msg.PosTimeInf)
	run.appLog.Printf("Setting postgreSQL -infinity time: %s", msg.NegTimeInf.Format(time.RFC3339Nano))
	run.appLog.Printf("Setting postgreSQL +infinity time: %s", msg.PosTimeInf.Format(time.RFC3339Nano))

	// connect to database
	run.appLog.Println(`Opening connection to postgreSQL database`)
	run.conn, err = sql.Open(driver, connect)
	if err != nil {
		run.appLog.Fatal(`Opening new database connection: `, err)
	}
	if err = run.conn.Ping(); err != nil {
		run.appLog.Fatal(`Testing new database connection: `, err)
	}
	run.dbConnected = true
	run.appLog.Println(`Database connection is alive`)

	run.appLog.Println(`Setting database connection timezone to: UTC`)
	if _, err = run.conn.Exec(stmt.DatabaseTimezone); err != nil {
		run.appLog.Fatal(`Setting session timezone: `, err)
	}

	run.appLog.Println(`Setting transaction isolation level to: SERIALIZABLE`)
	if _, err = run.conn.Exec(stmt.DatabaseIsolationLevel); err != nil {
		run.appLog.Fatal(`Setting transaction level: `, err)
	}

	// size the connection pool
	run.conn.SetMaxIdleConns(5)
	run.conn.SetMaxOpenConns(25)
	run.conn.SetConnMaxLifetime(12 * time.Hour)

	// required schema versions
	required := map[string]int64{
		`eye`: 201902120001,
	}

	// verify schema versions
	if rows, err = run.conn.Query(stmt.DatabaseSchemaVersion); err != nil {
		run.appLog.Fatal("Query db schema versions: ", err)
	}

rowloop:
	for rows.Next() {
		if err = rows.Scan(
			&schema,
			&version,
		); err != nil {
			run.appLog.Fatal(`DB schema check: `, err)
		}
		if rsv, ok := required[schema]; ok {
			if rsv != version {
				run.appLog.Fatalf("Incompatible schema %s: %d != %d", schema, rsv, version)
			}

			run.appLog.Printf("Detected DB schema %s, version: %d", schema, version)
			delete(required, schema)
			continue rowloop
		}
		run.appLog.Fatal(`Unknown registered schema: `, schema)
	}

	if err = rows.Err(); err != nil {
		run.appLog.Fatal(`DB schema check: `, err)
	}

	if len(required) != 0 {
		for s := range required {
			run.appLog.Printf("Missing database schema: %s", s)
		}
		run.appLog.Fatal(`DB schema check: incomplete database`)
	}
}

// pingDatabase continuously pings the database every second
func (run *runtime) pingDatabase() {
	ticker := time.NewTicker(time.Second).C

waitForConn:
	for {
		<-ticker
		if run.dbConnected {
			break waitForConn
		}
	}

	for {
		<-ticker
		err := run.conn.Ping()
		if err != nil {
			run.appLog.Warn(`main.runtime.pingDatabase: `, err)
		}
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
