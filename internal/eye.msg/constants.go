/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package msg // import "github.com/mjolnir42/eye/internal/eye.msg"

// Privileged access permission categories
const (
	CategoryOmnipotence = `omnipotence`
	CategorySystem      = `system`
)

// Section Supervisor handles AAA requests outside the permission model
const (
	SectionSupervisor  = `supervisor`
	ActionAuthenticate = `authenticate`
	TaskBasicAuth      = `basic-auth`
	VerdictOK          = 200
)

// Sections in category configuration are for action with configurations
// as scope
const (
	CategoryConfiguration = `configuration`
	SectionConfiguration  = `configuration`
)

// Actions for the various permission sections
const (
	ActionLookup = `lookup`
)

// Result codes
const (
	ResultOK = 200
)

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
