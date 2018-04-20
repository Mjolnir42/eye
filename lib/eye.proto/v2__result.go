/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto // import "github.com/mjolnir42/eye/lib/eye.proto"

import "net/http"

// Numeric result status codes used by eye
const (
	StatusOK             = 200
	StatusNoContent      = 204
	StatusPartial        = 206
	StatusBadRequest     = 400
	StatusUnauthorized   = 401
	StatusForbidden      = 403
	StatusNotFound       = 404
	StatusGone           = 410
	StatusUnprocessable  = 422
	StatusServerError    = 500
	StatusNotImplemented = 501
	StatusBadGateway     = 502
	StatusGatewayTimeout = 504
)

// Result ...
type Result struct {
	StatusCode     uint16           `json:"statusCode"`
	StatusText     string           `json:"statusText"`
	Errors         *[]string        `json:"errors,omitempty"`
	Configurations *[]Configuration `json:"configurations,omitempty"`
	Registrations  *[]Registration  `json:"registrations,omitempty"`
}

// SetStatus sets the status code
func (r *Result) SetStatus(code uint16) {
	r.StatusCode = code
	r.StatusText = http.StatusText(int(r.StatusCode))
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
