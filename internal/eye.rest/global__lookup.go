/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	msg "github.com/mjolnir42/eye/internal/eye.msg"
)

// LookupConfiguration function
func (x *Rest) LookupConfiguration(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionLookup
	request.Action = msg.ActionConfiguration
	request.LookupHash = strings.ToLower(params.ByName(`hash`))

	if !x.isAuthorized(&request) {
		x.replyForbidden(&w, &request, nil)
		return
	}

	// lookup is to be performed via SHA2/256 hash
	if len(request.LookupHash) != 64 {
		x.replyBadRequest(&w, &request, fmt.Errorf(
			`Invalid SHA2-256 lookup hash format`,
		))
		return
	}

	handler := x.handlerMap.Get(`lookup_r`)
	handler.Intake() <- request
	result := <-request.Reply
	x.respond(&w, &result)
}

// LookupRegistration accepts lookup requests for all registrations of a
// specific application. Internally this is mapped as a special case on
// top of RegistrationSearch.
func (x *Rest) LookupRegistration(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionLookup
	request.Action = msg.ActionRegistration
	request.Search.Registration.Application = params.ByName(`application`)

	if !x.isAuthorized(&request) {
		x.replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`registration_r`)
	handler.Intake() <- request
	result := <-request.Reply
	x.respond(&w, &result)
}

// LookupActivation accepts lookup requests for all activated
// configurations.
func (x *Rest) LookupActivation(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionLookup
	request.Action = msg.ActionActivation

	// parse URL query parameters to differentiate between
	// ActionActivation
	// and ActionSearch. Any number of parameters can be specified at
	// the same time
	if err := r.ParseForm(); err != nil {
		x.replyBadRequest(&w, &request, err)
		return
	}

	// Check if this an incremental update request for activations after
	// a specific time
	if activatedSince := r.Form.Get(`since`); activatedSince != `` {
		if err := stringToTime(
			activatedSince,
			&request.Search.Since,
		); err != nil {
			x.replyBadRequest(&w, &request, err)
			return
		}
	}
	// since -infinity == list all
	if request.Search.Since.IsZero() {
		request.Search.Since = msg.NegTimeInf
	}

	// Check if this is a request for pending activations instead of
	// active activations
	if pending := r.Form.Get(`pending`); pending != `` {
		var val bool
		var err error
		if val, err = strconv.ParseBool(pending); err != nil {
			x.replyBadRequest(&w, &request, err)
			return
		}
		request.Action = msg.ActionPending
		request.Flags.Pending = val
	}

	if !x.isAuthorized(&request) {
		x.replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`lookup_r`)
	handler.Intake() <- request
	result := <-request.Reply
	x.respond(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
