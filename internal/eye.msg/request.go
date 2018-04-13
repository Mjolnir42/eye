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
	uuid "github.com/satori/go.uuid"
)

// Request represents the internal request metadata
type Request struct {
	ID         uuid.UUID
	RemoteAddr string
	AuthUser   string
	Reply      chan Result
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
