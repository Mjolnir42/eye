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

// RegistrationList accepts requests to list all registrations
func (x *Rest) RegistrationList(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionRegistration
	request.Action = msg.ActionList

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
