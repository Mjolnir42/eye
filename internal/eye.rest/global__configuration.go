/*-
 * Copyright (c) 2016, Jörg Pernfuß <code.jpe@gmail.com>
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

// ConfigurationShow accepts requests to retrieve a specific
// configuration
func (x *Rest) ConfigurationShow(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionConfiguration
	request.Action = msg.ActionShow
	request.Configuration.ID = strings.ToLower(params.ByName(`ID`))

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	if _, err := uuid.FromString(request.Configuration.ID); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	handler := x.handlerMap.Get(`configuration_r`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// ConfigurationList accepts requests to list all configurations
func (x *Rest) ConfigurationList(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionConfiguration
	request.Action = msg.ActionList

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`configuration_r`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// ConfigurationAdd accepts requests to add a configuration
func (x *Rest) ConfigurationAdd(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionConfiguration
	request.Action = msg.ActionAdd

	cReq := proto.NewConfigurationRequest()
	if err := decodeJSONBody(r, &cReq); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}
	request.Configuration = *cReq.Configuration
	request.LookupHash = calculateLookupID(
		request.Configuration.HostID,
		request.Configuration.Metric,
	)

	if err := resolveFlags(&cReq, &request); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`configuration_w`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// ConfigurationUpdate accepts requests to update a configuration
func (x *Rest) ConfigurationUpdate(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionConfiguration
	request.Action = msg.ActionUpdate

	cReq := proto.NewConfigurationRequest()
	if err := decodeJSONBody(r, &cReq); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}
	request.Configuration = *cReq.Configuration
	request.LookupHash = calculateLookupID(
		request.Configuration.HostID,
		request.Configuration.Metric,
	)

	if request.Configuration.ID != params.ByName(`ID`) {
		replyBadRequest(&w, &request, fmt.Errorf(
			"Mismatched IDs in update: [%s] vs [%s]",
			request.Configuration.ID,
			params.ByName(`ID`),
		))
	}

	if err := resolveFlags(&cReq, &request); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`configuration_w`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// ConfigurationRemove accepts requests to remove a configuration
func (x *Rest) ConfigurationRemove(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionConfiguration
	request.Action = msg.ActionRemove
	request.Configuration.ID = params.ByName(`ID`)

	// request body may contain request flag overrides
	cReq := proto.NewConfigurationRequest()
	if err := decodeJSONBody(r, &cReq); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	if err := resolveFlags(&cReq, &request); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`configuration_w`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// ConfigurationActivate accepts requests to activate a configuration
func (x *Rest) ConfigurationActivate(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionConfiguration
	request.Action = msg.ActionActivate
	request.Configuration.ID = params.ByName(`ID`)

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`configuration_w`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
