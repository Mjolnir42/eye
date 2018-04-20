/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package msg // import "github.com/mjolnir42/eye/internal/eye.msg"

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	proto "github.com/mjolnir42/eye/lib/eye.proto"
	uuid "github.com/satori/go.uuid"
)

// Request represents the internal request metadata
type Request struct {
	ID           uuid.UUID
	Section      string
	Action       string
	RemoteAddr   string
	AuthUser     string
	Super        Supervisor
	Reply        chan Result
	LookupHash   string
	FeedbackURL  string
	Notification struct {
		ID         uuid.UUID
		PathPrefix string
	}

	Flags Flags

	ConfigurationTask string
	Configuration     proto.Configuration
	Registration      proto.Registration
}

// Flags represents the fully resolved proto.Request flags as they
// should be applied to this request
type Flags struct {
	AlarmClearing          bool
	CacheInvalidation      bool
	SendDeploymentFeedback bool
}

// New returns a Request
func New(r *http.Request, params httprouter.Params) Request {
	returnChannel := make(chan Result, 1)
	return Request{
		ID:         requestID(params),
		RemoteAddr: remoteAddr(r),
		AuthUser:   authUser(params),
		Reply:      returnChannel,
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
