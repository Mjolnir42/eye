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
	proto "github.com/mjolnir42/eye/lib/eye.proto"
	uuid "github.com/satori/go.uuid"
)

// RegistrationShow accepts requests to retrieve a specific registration
func (x *Rest) RegistrationShow(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionRegistration
	request.Action = msg.ActionShow
	request.Registration.ID = strings.ToLower(params.ByName(`ID`))

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	if _, err := uuid.FromString(request.Registration.ID); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	handler := x.handlerMap.Get(`registration_r`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// RegistrationList accepts requests to list all registrations. If r
// contains URL query parameters that indicate a search request, the
// returned list will be filtered for those search terms
func (x *Rest) RegistrationList(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionRegistration
	request.Action = msg.ActionList

	// parse URL query parameters to differentiate between ActionList
	// and ActionSearch. Any number of parameters can be specified at
	// the same time
	if err := r.ParseForm(); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}
	if app := r.Form.Get(`application`); app != `` {
		request.Action = msg.ActionSearch
		request.Search.Registration.Application = app
	}
	if addr := r.Form.Get(`address`); addr != `` {
		request.Action = msg.ActionSearch
		request.Search.Registration.Address = addr
	}
	if port := r.Form.Get(`port`); port != `` {
		if iPort, err := strconv.ParseInt(port, 10, 64); err == nil {
			request.Search.Registration.Port = iPort
		} else {
			replyBadRequest(&w, &request, err)
			return
		}
		request.Action = msg.ActionSearch
	}
	if db := r.Form.Get(`database`); db != `` {
		if iDb, err := strconv.ParseInt(db, 10, 64); err == nil {
			request.Search.Registration.Database = iDb
		} else {
			replyBadRequest(&w, &request, err)
			return
		}
		request.Action = msg.ActionSearch
	}

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`registration_r`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// RegistrationAdd accepts requests to add a registration
func (x *Rest) RegistrationAdd(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionRegistration
	request.Action = msg.ActionAdd

	cReq := proto.NewRegistrationRequest()
	if err := decodeJSONBody(r, &cReq); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}
	request.Registration = *cReq.Registration

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`registration_w`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// RegistrationUpdate accepts requests to update a registration
func (x *Rest) RegistrationUpdate(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionRegistration
	request.Action = msg.ActionUpdate

	cReq := proto.NewRegistrationRequest()
	if err := decodeJSONBody(r, &cReq); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}
	request.Registration = *cReq.Registration

	if request.Registration.ID != params.ByName(`ID`) {
		replyBadRequest(&w, &request, fmt.Errorf(
			"Mismatched IDs in update: [%s] vs [%s]",
			request.Registration.ID,
			params.ByName(`ID`),
		))
	}

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`registration_w`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// RegistrationRemove accepts requests to remove a registration
func (x *Rest) RegistrationRemove(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionRegistration
	request.Action = msg.ActionRemove
	request.Registration.ID = params.ByName(`ID`)

	// request body may contain request flag overrides
	cReq := proto.NewRegistrationRequest()
	if err := decodeJSONBody(r, &cReq); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`registration_w`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
