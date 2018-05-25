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
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mjolnir42/eye/lib/eye.proto/v2"
	uuid "github.com/satori/go.uuid"
)

// Request represents the internal request metadata
type Request struct {
	ID           uuid.UUID
	Time         time.Time
	Section      string
	Action       string
	Version      int
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

	Flags  Flags
	Search Search

	ConfigurationTask string
	Configuration     v2.Configuration
	Registration      v2.Registration
}

// Flags represents the fully resolved proto.Request flags as they
// should be applied to this request
type Flags struct {
	AlarmClearing          bool
	CacheInvalidation      bool
	SendDeploymentFeedback bool
	ResetActivation        bool
}

// Search contains search paramaters for this request
type Search struct {
	Registration  v2.Registration
	Configuration v2.Configuration
	ValidAt       time.Time
}

// New returns a Request
func New(r *http.Request, params httprouter.Params) Request {
	returnChannel := make(chan Result, 1)
	var protocolVersion int
	switch {
	case strings.HasPrefix(r.URL.EscapedPath(), `/api/v1/`):
		protocolVersion = ProtocolOne
	case strings.HasPrefix(r.URL.EscapedPath(), `/api/v2/`):
		protocolVersion = ProtocolTwo
	}
	return Request{
		ID:         requestID(params),
		Time:       requestTS(params),
		RemoteAddr: remoteAddr(r),
		AuthUser:   authUser(params),
		Reply:      returnChannel,
		Version:    protocolVersion,
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
