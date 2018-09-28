/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package msg // import "github.com/solnx/eye/internal/eye.msg"

// API protocol versions
const (
	ProtocolInvalid int = iota
	ProtocolOne
	ProtocolTwo
)

// Privileged access permission categories
const (
	CategoryOmnipotence = `omnipotence`
	CategorySystem      = `system`
)

// Section Supervisor handles AAA requests outside the permission model
const (
	SectionSupervisor   = `supervisor`
	TaskBasicAuth       = `basic-auth`
	VerdictOK           = 200
	VerdictUnauthorized = 401
)

// Sections in category global are unscoped sections
const (
	CategoryGlobal       = `global`
	SectionConfiguration = `configuration`
	SectionDeployment    = `deployment`
	SectionLookup        = `lookup`
	SectionRegistration  = `registration`
	TaskClearing         = `clearing`
	TaskDelete           = `delete`
	TaskDeprovision      = `deprovision`
	TaskPending          = `pending`
	TaskRollout          = `rollout`
	TaskUpdate           = `update`
)

// Actions for the various permission sections
const (
	ActionActivate      = `activate`
	ActionActivation    = `activation`
	ActionAdd           = `add`
	ActionAuthenticate  = `authenticate`
	ActionAuthorize     = `authorize`
	ActionConfiguration = `configuration`
	ActionHistory       = `history`
	ActionList          = `list`
	ActionNop           = `nop`
	ActionNotification  = `notification`
	ActionPending       = `pending`
	ActionProcess       = `process`
	ActionRegistration  = `registration`
	ActionRemove        = `remove`
	ActionSearch        = `search`
	ActionShow          = `show`
	ActionUpdate        = `update`
	ActionVersion       = `version`
)

// Result codes
const (
	ResultOK             = 200
	ResultNoContent      = 204
	ResultBadRequest     = 400
	ResultUnauthorized   = 401
	ResultForbidden      = 403
	ResultNotFound       = 404
	ResultGone           = 410
	ResultUnprocessable  = 422
	ResultServerError    = 500
	ResultNotImplemented = 501
	ResultBadGateway     = 502
	ResultGatewayTimeout = 504
)

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
