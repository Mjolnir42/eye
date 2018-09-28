/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

// Package mock implements various placeholders
package mock // import "github.com/solnx/eye/internal/eye.mock"

import (
	msg "github.com/solnx/eye/internal/eye.msg"
)

// AlwaysAuthorize authorizes every request
func AlwaysAuthorize(q *msg.Request) bool {
	return true
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
