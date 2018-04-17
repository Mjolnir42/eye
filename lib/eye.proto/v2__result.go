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
	StatusOK      = 200
	StatusPartial = 206
)

// DisplayStatus holds the string representation of the various status
// codes
var DisplayStatus = map[int]string{
	200: "OK",
	206: "Partial result",
}

// Result ...
type Result struct {
	StatusCode     uint16           `json:"statusCode"`
	StatusText     string           `json:"statusText"`
	Errors         *[]string        `json:"errors,omitempty"`
	Configurations *[]Configuration `json:"configurations,omitempty"`
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

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
