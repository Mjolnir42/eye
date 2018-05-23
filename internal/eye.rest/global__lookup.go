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
		replyForbidden(&w, &request, nil)
		return
	}

	// lookup is to be performed via SHA2/256 hash
	if len(request.LookupHash) != 64 {
		replyBadRequest(&w, &request, fmt.Errorf(
			`Invalid SHA2-256 lookup hash format`,
		))
		return
	}

	handler := x.handlerMap.Get(`lookup_r`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
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
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`registration_r`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
