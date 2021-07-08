/*
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt // import "github.com/solnx/eye/internal/eye.stmt"

// ProvisionStatements contains the SQL statements related to
// provisioning records in Eye
const (
	ProvisionStatements = ``

	ProvAdd = `
INSERT INTO eye.provisions (
            dataID,
            configurationID,
            provision_period,
            tasks
)
SELECT $1::uuid,
       $2::uuid,
       tstzrange($3::timestamptz, 'infinity', '[]'),
       $4::varchar[]
WHERE  NOT EXISTS (
       SELECT dataID
       FROM   eye.provisions
       WHERE  dataID = $1::uuid);`

	ProvFinalize = `
UPDATE eye.provisions
SET    provision_period = tstzrange(( SELECT lower(provision_period)
                                      FROM   eye.provisions
                                      WHERE  dataID = $1::uuid ),
                                    $2::timestamptz,
                                    '[]'),
       tasks = array_append(( SELECT tasks
                              FROM   eye.provisions
                              WHERE  dataID = $1::uuid ),
                            $3::varchar)
WHERE  dataID = $1::uuid;`

	ProvForDataID = `
SELECT lower(provision_period),
       upper(provision_period),
       tasks
FROM   eye.provisions
WHERE  dataID = $1::uuid;`
)

func init() {
	m[ProvAdd] = `ProvAdd`
	m[ProvFinalize] = `ProvFinalize`
	m[ProvForDataID] = `ProvForDataID`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
