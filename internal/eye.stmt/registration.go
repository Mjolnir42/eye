/*
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt // import "github.com/mjolnir42/eye/internal/eye.stmt"

// RegistryStatements contains the SQL statements related to the
// registration of applications with Eye
const (
	RegistryStatements = ``

	RegistryCreateTable = `
CREATE TABLE IF NOT EXISTS eye.registry (
  registrationID          uuid            PRIMARY KEY,
  application             varchar(128)    NOT NULL,
  address                 inet            NOT NULL,
  port                    numeric(5,0)    NOT NULL CONSTRAINT valid_port CHECK ( port > 0 AND port < 65536 ),
  database                numeric(5,0)    NOT NULL CONSTRAINT valid_db CHECK ( database >= 0 ),
  registeredAt            timestamptz(3)  NOT NULL DEFAULT NOW(),
  CONSTRAINT registeredAt_utc CHECK( EXTRACT( TIMEZONE FROM registeredAt ) = '0' ),
  UNIQUE( application, address, port, database )
);`

	RegistryAdd = `
INSERT INTO eye.registry (
            registrationID,
            application,
            address,
            port,
            database)
SELECT $1::uuid,
       $2::varchar,
       $3::inet,
       $4::numeric,
       $5::numeric
WHERE  NOT EXISTS (
       SELECT registrationID
       FROM   eye.registry
       WHERE  application = $2::varchar
         AND  address     = $3::inet
         AND  port        = $4::numeric
         AND  database    = $5::numeric);`

	RegistryDel = `
DELETE FROM eye.registry
WHERE  registrationID = $1::uuid;`

	RegistrySearch = `
SELECT registrationID
FROM   eye.registry
WHERE  application = $1::varchar
  AND  address = $2::inet
  AND  port = $3::numeric
  AND  database = $4::numeric;`

	RegistryList = `
SELECT registrationID
FROM   eye.registry;`

	RegistryShow = `
SELECT registrationID,
       application,
       address,
       port,
       database,
       registeredAt
FROM   eye.registry
WHERE  registrationID = $1::uuid;`
)

func init() {
	m[RegistryAdd] = `RegistryAdd`
	m[RegistryCreateTable] = `RegistryCreateTable`
	m[RegistryDel] = `RegistryDel`
	m[RegistryList] = `RegistryList`
	m[RegistrySearch] = `RegistrySearch`
	m[RegistryShow] = `RegistryShow`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
