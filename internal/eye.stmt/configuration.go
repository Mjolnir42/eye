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

	CfgList = `
SELECT configurationID
FROM   eye.configurations_data
WHERE  validity @> NOW()::timestamptz;`

	ConfigurationExists = `
SELECT configurationID
FROM   eye.configurations
WHERE  configurationID = $1::uuid;`

	ConfigurationUpdate = `
UPDATE eye.configurations
SET    lookupID = $2::varchar,
       configuration = $3::jsonb
WHERE  configurationID = $1::uuid;`

	CfgAddID = `
INSERT INTO eye.configurations (
            configurationID,
            lookupID
)
SELECT $1::uuid,
       $2::varchar
WHERE  NOT EXISTS (
       SELECT configurationID
       FROM   eye.configurations
       WHERE  configurationID = $1::uuid);`

	CfgSelectValidForUpdate = `
SELECT dataID,
       lower(validity)
FROM   eye.configurations_data
WHERE  configurationID = $1::uuid
  AND  validity @> NOW()::timestamptz
FOR    UPDATE;`

	CfgSelectValid = `
SELECT dataID,
       lower(validity)
FROM   eye.configurations_data
WHERE  configurationID = $1::uuid
  AND  validity @> NOW()::timestamptz;`

	CfgDataUpdateValidity = `
UPDATE eye.configurations_data
SET    validity = tstzrange($1::timestamptz, $2::timestamptz, '[)')
WHERE  dataID = $3::uid;`

	CfgAddData = `
INSERT INTO eye.configurations_data (
            dataID,
            configurationID,
            validity,
            configuration
)
SELECT $1::uuid,
       $2::uuid,
       tstzrange($3::timestamptz, 'infinity', '[]'),
       $4::jsonb
WHERE  NOT EXISTS (
       SELECT dataID
       FROM   eye.configurations_data
       WHERE  dataID = $1::uuid);`

	CfgShow = `
SELECT d.configuration,
       upper(d.validity),
       lower(p.provision_period),
       upper(p.provision_period),
       p.tasks
FROM   eye.configurations_data AS d
JOIN   eye.provisions AS p
  ON   d.dataID = p.dataID
WHERE  d.dataID = $1::uuid;`
)

func init() {
	m[ConfigurationExists] = `ConfigurationExists`
	m[ConfigurationUpdate] = `ConfigurationUpdate`

	m[CfgAddID] = `CfgAddID`
	m[CfgSelectValidForUpdate] = `CfgSelectValidForUpdate`
	m[CfgSelectValid] = `CfgSelectValid`
	m[CfgDataUpdateValidity] = `CfgDataUpdateValidity`
	m[CfgAddData] = `CfgAddData`
	m[CfgShow] = `CfgShow`
	m[CfgList] = `CfgList`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
