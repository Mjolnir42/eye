/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/solnx/eye/internal/eye.rest"

import (
	"database/sql"

	msg "github.com/solnx/eye/internal/eye.msg"
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

func (x *Rest) updateRegister() error {
	var (
		rows                                    *sql.Rows
		err                                     error
		registrationID, address, port, database string
	)
	// perform search
	if rows, err = x.stmtRegisterGetAll.Query(); err != nil {
		x.appLog.Errorf("Section=%s Action=%s Error=%s", "Invalidation/Rest", "UpdateRegister", err.Error())
		return err
	}
	tmpMap := make(map[string][3]string)
	// iterate over result list
	for rows.Next() {
		if err = rows.Scan(
			&registrationID,
			&address,
			&port,
			&database,
		); err != nil {
			rows.Close()
			x.appLog.Errorf("Section=%s Action=%s Error=%s", "Invalidation/Rest", "UpdateRegister", err.Error())
			return err
		}
		// build result list

		tmpMap[registrationID] = [3]string{address, port, database}
	}
	return x.invl.UpdateAll(tmpMap)
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
	//Update Register of redis servers from database
	x.updateRegister()

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
invalidation_loop:
	for {
		select {
		case <-errors:
		case <-done:
			break invalidation_loop
		}
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
