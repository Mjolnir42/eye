/*
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt // import "github.com/mjolnir42/eye/internal/eye.stmt"

// DatabaseStatements contains statements related to managing and
// configuring the database connections
const (
	DatabaseStatements = ``

	DatabaseTimezone = `SET TIME ZONE 'UTC';`

	DatabaseIsolationLevel = `SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE;`

	DatabaseSchemaVersion = `
SELECT schema,
       MAX(version) AS version
FROM   public.schema_versions
GROUP  BY schema;`
)

func init() {
	m[DatabaseTimezone] = `DatabaseTimezone`
	m[DatabaseIsolationLevel] = `DatabaseIsolationLevel`
	m[DatabaseSchemaVersion] = `DatabaseSchemaVersion`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
