/*
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt // import "github.com/mjolnir42/eye/internal/eye.stmt"

// ConfigurationStatements contains the SQL statements related to
// configurations in Eye
const (
	ConfigurationStatements = ``

	ConfigurationList = `
SELECT configurationID
FROM   eye.configurations;`

	ConfigurationShow = `
SELECT configuration
FROM   eye.configurations
WHERE  configurationID = $1::uuid;`

	ConfigurationExists = `
SELECT configurationID
FROM   eye.configurations
WHERE  configurationID = $1::uuid;`

	ConfigurationAdd = `
INSERT INTO eye.configurations (
            configurationID,
            lookupID,
            configuration)
SELECT $1::uuid,
       $2::varchar,
       $3::jsonb
WHERE  NOT EXISTS (
       SELECT configurationID
       FROM   eye.configurations
       WHERE  configurationID = $1::uuid);`

	ConfigurationRemove = `
DELETE FROM eye.configurations
WHERE       configurationID = $1::uuid;`

	ConfigurationUpdate = `
UPDATE eye.configurations
SET    lookupID = $2::varchar,
       configuration = $3::jsonb
WHERE  configurationID = $1::uuid;`

	ConfigurationCountForLookupID = `
SELECT COUNT(1)::integer
FROM   eye.configurations
WHERE  lookupID = $1::varchar;`

	ConfigurationActivate = `
INSERT INTO eye.activations (
            configurationID)
SELECT $1::uuid
WHERE  NOT EXISTS (
       SELECT configurationID
       FROM   eye.activations
       WHERE  configurationID = $1::uuid;`

	ConfigurationProvision = `
INSERT INTO eye.provisions (
            configurationID)
SELECT $1::uuid
WHERE  NOT EXISTS (
       SELECT configurationID
       FROM   eye.provisions
       WHERE  configurationID = $1::uuid;`
)

func init() {
	m[ConfigurationActivate] = `ConfigurationActivate`
	m[ConfigurationAdd] = `ConfigurationAdd`
	m[ConfigurationCountForLookupID] = `ConfigurationCountForLookupID`
	m[ConfigurationExists] = `ConfigurationExists`
	m[ConfigurationList] = `ConfigurationList`
	m[ConfigurationProvision] = `ConfigurationProvision`
	m[ConfigurationRemove] = `ConfigurationRemove`
	m[ConfigurationShow] = `ConfigurationShow`
	m[ConfigurationUpdate] = `ConfigurationUpdate`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
