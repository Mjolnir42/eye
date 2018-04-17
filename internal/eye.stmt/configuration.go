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
)

func init() {
	m[ConfigurationList] = `ConfigurationList`
	m[ConfigurationShow] = `ConfigurationShow`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
