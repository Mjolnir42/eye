/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import (
	msg "github.com/mjolnir42/eye/internal/eye.msg"
)

// eyewallCacheRegister adds a cache to the invalidation registry
func (x *Rest) eyewallCacheRegister(r *msg.Result) {
	switch r.Section {
	case msg.SectionRegistration:
	default:
		return
	}

	switch r.Action {
	case msg.ActionAdd:
	case msg.ActionUpdate:
	default:
		return
	}

	switch r.Code {
	case msg.ResultOK:
	default:
		return
	}

	reg := r.Registration[0]
	x.invl.Register(reg.ID, reg.Address, reg.Port, reg.Database)
}

// eyewallCacheUnregister removes a cache from the invalidation registry
func (x *Rest) eyewallCacheUnregister(r *msg.Result) {
	switch r.Section {
	case msg.SectionRegistration:
	default:
		return
	}

	switch r.Action {
	case msg.ActionRemove:
	case msg.ActionUpdate:
	default:
		return
	}

	switch r.Code {
	case msg.ResultOK:
	default:
		return
	}

	reg := r.Registration[0]
	x.invl.Unregister(reg.ID)
}

// eyewallCacheInvalidate triggers cache invalidation for results that
// support it
func (x *Rest) eyewallCacheInvalidate(r *msg.Result) {
	if !r.Flags.CacheInvalidation {
		return
	}

	switch r.Section {
	case msg.SectionConfiguration:
	case msg.SectionDeployment:
	default:
		return
	}

	switch r.Action {
	case msg.ActionAdd:
	case msg.ActionUpdate:
	case msg.ActionRemove:
	case msg.ActionNotification:
	case msg.ActionProcess:
	default:
		return
	}

	if !r.Flags.AlarmClearing {
		// asynchronous active cache invalidation, since no
		// clearing action depends on the invalidation having been
		// performed
		x.invl.AsyncInvalidate(r.Configuration[0].LookupID)
		return
	}

	// r.Flags.AlarmClearing == true

	// synchronous active cache invalidation, since the
	// clearing has to be blocked until the invalidation has been
	// performed
	done, errors := x.invl.Invalidate(r.Configuration[0].LookupID)
	for {
		select {
		case <-errors:
		case <-done:
			break
		}
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
