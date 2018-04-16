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
FROM   eye.configuration_items
WHERE  lookup_id = $1::varchar;`
)

func init() {
	m[LookupSearch] = `LookupSearch`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
