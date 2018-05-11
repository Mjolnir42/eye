/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import (
	"net/http"

	msg "github.com/mjolnir42/eye/internal/eye.msg"
)

// replyNoContent returns a 204 result
func replyNoContent(w *http.ResponseWriter, q *msg.Request) {
	result := msg.FromRequest(q)
	result.NoContent()
	respond(w, &result)
}

// replyBadRequest returns a 400 error
func replyBadRequest(w *http.ResponseWriter, q *msg.Request, err error) {
	result := msg.FromRequest(q)
	result.BadRequest(err)
	respond(w, &result)
}

// replyForbidden returns a 403 error
func replyForbidden(w *http.ResponseWriter, q *msg.Request, err error) {
	result := msg.FromRequest(q)
	result.Forbidden(err)
	respond(w, &result)
}

// replyGone returns a 410 error
func replyGone(w *http.ResponseWriter, q *msg.Request, err error) {
	result := msg.FromRequest(q)
	result.Gone(err)
	respond(w, &result)
}

// replyUnprocessableEntity returns a 422 error
func replyUnprocessableEntity(w *http.ResponseWriter, q *msg.Request, err error) {
	result := msg.FromRequest(q)
	result.UnprocessableEntity(err)
	respond(w, &result)
}

// replyInternalError returns a 500 error
func replyInternalError(w *http.ResponseWriter, q *msg.Request, err error) {
	result := msg.FromRequest(q)
	result.ServerError(err)
	respond(w, &result)
}

// replyBadGateway returns a 502 error
func replyBadGateway(w *http.ResponseWriter, q *msg.Request, err error) {
	result := msg.FromRequest(q)
	result.BadGateway(err)
	respond(w, &result)
}

// replyGatewayTimeout returns a 504 error
func replyGatewayTimeout(w *http.ResponseWriter, q *msg.Request, err error) {
	result := msg.FromRequest(q)
	result.GatewayTimeout(err)
	respond(w, &result)
}

// sendJSONReply returns a 200 status JSON result
func sendJSONReply(w *http.ResponseWriter, b *[]byte) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusOK)
	(*w).Write(*b)
}

// sendV1Result returns API Protocol version 1 results
func sendV1Result(w *http.ResponseWriter, code uint16, errstr string, body *[]byte) {
	if errstr != `` {
		http.Error(*w, errstr, int(code))
		return
	}
	if body != nil {
		(*w).Header().Set("Content-Type", "application/json")
	}
	(*w).WriteHeader(int(code))
	if body == nil {
		(*w).Write(nil)
		return
	}
	(*w).Write(*body)
}

// hardInternalError returns a 500 server error with no application data
// body. This function is intended to be used only if normal response
// generation itself fails
func hardInternalError(w *http.ResponseWriter) {
	http.Error(*w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
