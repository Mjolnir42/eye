/*
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt // import "github.com/solnx/eye/internal/eye.stmt"

// ActivationStatements contains the SQL statements related to
// recording configurations as activated in Eye
const (
	ActivationStatements = ``

	ActivationGet = `
SELECT activatedAt
FROM   eye.activations
WHERE  configurationID = $1::uuid;`

	ActivationDel = `
DELETE FROM eye.activations
WHERE  configurationID = $1::uuid;`

	ActivationSet = `
INSERT INTO eye.activations (
            configurationID)
SELECT $1::uuid
WHERE  NOT EXISTS (
       SELECT configurationID
       FROM   eye.activations
       WHERE  configurationID = $1::uuid);`
)

func init() {
	m[ActivationDel] = `ActivationDel`
	m[ActivationGet] = `ActivationGet`
	m[ActivationSet] = `ActivationSet`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
