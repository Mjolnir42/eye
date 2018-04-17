/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package msg // import "github.com/mjolnir42/eye/internal/eye.msg"

import (
	"fmt"

	proto "github.com/mjolnir42/eye/lib/eye.proto"
	uuid "github.com/satori/go.uuid"
)

// Result ...
type Result struct {
	ID      uuid.UUID
	Section string
	Action  string
	Code    uint16
	Error   error
	Super   Supervisor

	FeedbackURL       string
	ConfigurationTask string
	Configuration     []proto.Configuration

	fixated bool
}

// FromRequest returns a Result configured to match Request rq
func FromRequest(rq *Request) Result {
	return Result{
		ID:                rq.ID,
		Section:           rq.Section,
		Action:            rq.Action,
		FeedbackURL:       rq.FeedbackURL,
		ConfigurationTask: rq.ConfigurationTask,
	}
}

// UnknownRequest is a wrapper function for NotImplemented using a
// default error based on Request q
func (r *Result) UnknownRequest(q *Request) {
	r.NotImplemented(fmt.Errorf(
		"Unknown requested action: %s/%s",
		q.Section,
		q.Action,
	))
}

// OK configures the result to reflect that the request was processed
// fully and without error
func (r *Result) OK() {
	r.shrinkwrap(ResultOK, nil)
}

// NotFound configures the result to reflect that the request target was
// not found
func (r *Result) NotFound(err error) {
	r.shrinkwrap(ResultNotFound, err)
}

// ServerError configures the result to reflect an occurred server error
func (r *Result) ServerError(err error) {
	r.shrinkwrap(ResultServerError, err)
}

// NotImplemented configures the result to reflect that a codepath was
// requested that is not implemented
func (r *Result) NotImplemented(err error) {
	r.shrinkwrap(ResultNotImplemented, err)
}

// HasFailed returns true if the Result r is for a request that has
// failed. If the result code has not been set, the result is considered
// failed as well.
func (r *Result) HasFailed() bool {
	if r.Code == 0 || r.Code > 299 {
		return true
	}
	return false
}

// shrinkwrap finalizes the Result r
func (r *Result) shrinkwrap(code uint16, err error) {
	if r.fixated {
		assertIsNil(fmt.Errorf("msg: double-shrinkwrap of result for RequestID %s",
			r.ID.String(),
		))
	}
	r.Code = code
	r.setError(err)
	r.clear()
	r.fixated = true
}

// setError sets r.Error to err, unless err is nil in which case r.Error
// is preserved as is
func (r *Result) setError(err error) {
	if err != nil {
		r.Error = err
	}
}

// clear wipes the result from potential partial results
func (r *Result) clear() {
	switch r.Section {
	case SectionLookup:
		r.Configuration = []proto.Configuration{}
	case SectionDeployment:
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
