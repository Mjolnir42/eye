/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto // import "github.com/solnx/eye/lib/eye.proto"

// API protocol versions
const (
	ProtocolInvalid int = iota
	ProtocolOne
	ProtocolTwo
)

// RFC3339Milli is a millisecond precision RFC3339 timestamp format
// definition
const RFC3339Milli string = "2006-01-02T15:04:05.000Z07:00"

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
