/*
 * Copyright (c) 2016, 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt // import "github.com/mjolnir42/eye/internal/eye.stmt"

// LookupStatements contains the SQL statements related to lookup
// requests
const (
	LookupStatements = ``

	LookupConfiguration = `
SELECT    c.configurationID,
          d.dataID,
          lower(d.validity),
          upper(d.validity),
          d.configuration,
		  lower(p.provision_period),
		  upper(p.provision_period),
		  p.tasks,
	      a.activatedAt
FROM      eye.configurations AS c
JOIN      eye.configurations_data AS d
  ON      c.configurationID = d.configurationID
JOIN      eye.provisions AS p
  ON      d.dataID = p.dataID
LEFT JOIN eye.activations AS a
       ON c.configurationID = a.configurationID
WHERE     c.lookupID = $1::varchar
  AND     d.validity @> NOW()::timestamptz;`

	LookupAddID = `
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

	LookupActivation = `
SELECT a.configurationID,
       a.activatedAt,
       c.lookupID,
       d.dataID,
       lower(d.validity),
       upper(d.validity),
       configuration
FROM   eye.activations a
JOIN   eye.configurations c
  ON   a.configurationID = c.configurationID
JOIN   eye.configurations_data d
  ON   c.configurationID = d.configurationID
WHERE  d.validity @> NOW()::timestamptz
  AND  a.activatedAt >= $1::timestamptz;`

	LookupPending = `
SELECT     p.configurationID,
           lower(p.provision_period),
           d.configuration
FROM       eye.provisions p
LEFT OUTER
      JOIN eye.activations a
        ON p.configurationID = a.configurationID
JOIN       eye.configurations_data d
        ON p.dataID = d.dataID
       AND p.configurationID = d.configurationID
WHERE      p.provision_period @> NOW()::timestamptz
       AND a.configurationID IS NULL
       AND lower(p.provision_period) >= $1::timestamptz;`
)

func init() {
	m[LookupActivation] = `LookupActivation`
	m[LookupAddID] = `LookupAddID`
	m[LookupConfiguration] = `LookupConfiguration`
	m[LookupPending] = `LookupPending`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
