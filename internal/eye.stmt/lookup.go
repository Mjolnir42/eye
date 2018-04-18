/*
 * Copyright (c) 2016, 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt // import "github.com/mjolnir42/eye/internal/eye.stmt"

// LookupStatements contains the SQL statements related to configuration
// search via hash
const (
	LookupStatements = ``

	LookupSearch = `
SELECT configuration
FROM   eye.configurations
WHERE  lookupID = $1::varchar;`

	LookupExists = `
SELECT lookupID
FROM   eye.lookup
WHERE  lookupID = $1::varchar;`

	LookupAdd = `
INSERT INTO eye.lookup (
            lookupID,
            hostID,
            metric)
SELECT $1::varchar,
       $2::numeric,
       $3::text
WHERE  NOT EXISTS (
       SELECT lookupID
       FROM   eye.lookup
       WHERE  lookupID = $1::varchar
          OR  ( hostID = $2::numeric AND metric = $3::text));`

	LookupRemove = `
DELETE FROM eye.lookup
WHERE       lookupID = $1::varchar;`
)

func init() {
	m[LookupAdd] = `LookupAdd`
	m[LookupExists] = `LookupExists`
	m[LookupRemove] = `LookupRemove`
	m[LookupSearch] = `LookupSearch`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
