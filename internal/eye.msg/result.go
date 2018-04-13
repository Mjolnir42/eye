/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package msg // import "github.com/mjolnir42/eye/internal/eye.msg"

import (
	proto "github.com/mjolnir42/eye/lib/eye.proto"
	uuid "github.com/satori/go.uuid"
)

// Result ...
type Result struct {
	ID      uuid.UUID
	Section string
	Action  string
	Error   error
	Super   Supervisor

	Configuration []proto.Configuration
}

// FromRequest ...
func FromRequest(rq *Request) Result {
	return Result{
		ID:      rq.ID,
		Section: rq.Section,
		Action:  rq.Action,
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
