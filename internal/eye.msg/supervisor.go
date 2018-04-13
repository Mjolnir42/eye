/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package msg // import "github.com/mjolnir42/eye/internal/eye.msg"

// Supervisor contains data related to pending AAA operations
type Supervisor struct {
	Task      string
	Verdict   uint16
	BasicAuth struct {
		User  []byte
		Token []byte
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
