/*
 * Copyright (c) 2016, 1&1 Internet SE
 * Written by Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved.
 */

package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func connectToDatabase() {
	var err error
	driver := "postgres"

	connect := fmt.Sprintf(
		"%s='%s' %s='%s' %s='%s' %s='%s' %s='%s' %s='%s' %s='%s'",
		"dbname",
		Eye.Database.Name,
		"user",
		Eye.Database.User,
		"password",
		Eye.Database.Pass,
		"host",
		Eye.Database.Host,
		"port",
		Eye.Database.Port,
		"sslmode",
		Eye.Database.TLSMode,
		"connect_timeout",
		Eye.Database.Timeout,
	)

	Eye.run.conn, err = sql.Open(driver, connect)
	if err != nil {
		log.Fatal(err)
	}
	if err = Eye.run.conn.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Print("Connected to database")
	if _, err = Eye.run.conn.Exec(`SET TIME ZONE 'UTC';`); err != nil {
		log.Fatal(err)
	}
}

func pingDatabase() {
	ticker := time.NewTicker(time.Second).C

	for {
		<-ticker
		err := Eye.run.conn.Ping()
		if err != nil {
			log.Print(err)
		}
	}
}

func prepareStatements() {
	var err error

	Eye.run.getLookup, err = Eye.run.conn.Prepare(stmtGetLookupIDForItem)
	log.Println("Preparing: get_lookup")
	abortOnError(err)

	Eye.run.itemCount, err = Eye.run.conn.Prepare(stmtGetItemCountForLookupID)
	log.Println("Preparing: item_count")
	abortOnError(err)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
