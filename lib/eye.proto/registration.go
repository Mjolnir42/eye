/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto // import "github.com/mjolnir42/eye/lib/eye.proto"

import "time"

// Registration holds the information about a client component's
// cache registration
type Registration struct {
	ID           string    `json:"registrationID" valid:"uuidv4"`
	Application  string    `json:"application"`
	Address      string    `json:"address"`
	Port         int64     `json:"port,string"`
	Database     int64     `json:"database,string"`
	RegisteredAt time.Time `json:"registeredAt,string"`
}

// NewRegistrationRequest returns a new request
func NewRegistrationRequest() Request {
	return Request{
		Flags:        &Flags{},
		Registration: &Registration{},
	}
}

// NewRegistrationResult returns a new result
func NewRegistrationResult() Result {
	return Result{
		Errors:        &[]string{},
		Registrations: &[]Registration{},
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
