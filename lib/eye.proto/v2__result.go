/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto // import "github.com/mjolnir42/eye/lib/eye.proto"

// Numeric result status codes
const (
	StatusOK             = 200
	StatusPartial        = 206
	StatusServerError    = 500
	StatusNotImplemented = 501
)

// DisplayStatus holds the string representation of the various status
// codes
var DisplayStatus = map[int]string{
	200: "OK",
	206: "Partial result",
	500: "Server error",
	501: "Not implemented",
}

// Result ...
type Result struct {
	StatusCode     uint16           `json:"statusCode"`
	StatusText     string           `json:"statusText"`
	Errors         *[]string        `json:"errors,omitempty"`
	Configurations *[]Configuration `json:"configurations,omitempty"`
}

// SetStatus sets the status code
func (r *Result) SetStatus(code uint16) {
	switch code {
	case StatusOK:
		r.OK()
	case StatusPartial:
		r.Partial()
	case StatusServerError:
		r.ServerError()
	case StatusNotImplemented:
		r.NotImplemented()
	}
}

// OK updates the status fields to indicate a partial result if the
// result contains one or more error messages, and success otherwise
func (r *Result) OK() {
	if r.Errors == nil || *r.Errors == nil || len(*r.Errors) == 0 {
		r.StatusCode = StatusOK
		r.StatusText = DisplayStatus[StatusOK]
		return
	}
	r.Partial()
}

// Partial updates the status fields to indicate a partial result
func (r *Result) Partial() {
	r.StatusCode = StatusPartial
	r.StatusText = DisplayStatus[StatusPartial]
}

// ServerError ... 500
func (r *Result) ServerError() {
	r.StatusCode = StatusServerError
	r.StatusText = DisplayStatus[StatusServerError]
}

// NotImplemented ... 501
func (r *Result) NotImplemented() {
	r.StatusCode = StatusNotImplemented
	r.StatusText = DisplayStatus[StatusNotImplemented]
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
