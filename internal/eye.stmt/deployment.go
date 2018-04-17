/*
 * Copyright (c) 2016, 2018, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß <code.jpe@gmail.com>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt // import "github.com/mjolnir42/eye/internal/eye.stmt"

// DeploymentStatements contains the SQL statements related to
// configuration deployment
const (
	DeploymentStatements = ``

	ConfigurationExists = `
SELECT configurationID
FROM   eye.configurations
WHERE  configurationID = $1::uuid;`
)

func init() {
	m[ConfigurationExists] = `ConfigurationExists`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
