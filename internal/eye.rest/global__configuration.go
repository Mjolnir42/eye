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
	"time"

	"github.com/julienschmidt/httprouter"
	msg "github.com/mjolnir42/eye/internal/eye.msg"
	"github.com/mjolnir42/eye/lib/eye.proto/v1"
	"github.com/mjolnir42/eye/lib/eye.proto/v2"
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

	if _, err := uuid.FromString(request.Configuration.ID); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
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

	switch request.Version {
	case msg.ProtocolOne:
		cReq := &v1.ConfigurationItem{}
		if err := decodeJSONBody(r, cReq); err != nil {
			replyUnprocessableEntity(&w, &request, err)
			return
		}
		request.Configuration = v2.ConfigurationFromV1(cReq)

	case msg.ProtocolTwo:
		cReq := v2.NewConfigurationRequest()
		if err := decodeJSONBody(r, &cReq); err != nil {
			replyUnprocessableEntity(&w, &request, err)
			return
		}
		request.Configuration = *cReq.Configuration

		// only the v2 API has request flags
		if err := resolveFlags(&cReq, &request); err != nil {
			replyBadRequest(&w, &request, err)
			return
		}

	default:
		replyInternalError(&w, &request, nil)
		return
	}

	request.Configuration.InputSanatize()
	request.LookupHash = calculateLookupID(
		request.Configuration.HostID,
		request.Configuration.Metric,
	)
	request.Configuration.LookupID = request.LookupHash

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

	switch request.Version {
	case msg.ProtocolOne:
		cReq := &v1.ConfigurationItem{}
		if err := decodeJSONBody(r, cReq); err != nil {
			replyUnprocessableEntity(&w, &request, err)
			return
		}
		request.Configuration = v2.ConfigurationFromV1(cReq)

	case msg.ProtocolTwo:
		cReq := v2.NewConfigurationRequest()
		if err := decodeJSONBody(r, &cReq); err != nil {
			replyUnprocessableEntity(&w, &request, err)
			return
		}
		request.Configuration = *cReq.Configuration

		// only the v2 API has request flags
		if err := resolveFlags(&cReq, &request); err != nil {
			replyBadRequest(&w, &request, err)
			return
		}

	default:
		replyInternalError(&w, &request, nil)
		return
	}

	request.Configuration.InputSanatize()
	request.LookupHash = calculateLookupID(
		request.Configuration.HostID,
		request.Configuration.Metric,
	)
	request.Configuration.LookupID = request.LookupHash

	if request.Configuration.ID != strings.ToLower(params.ByName(`ID`)) {
		replyBadRequest(&w, &request, fmt.Errorf(
			"Mismatched IDs in update: [%s] vs [%s]",
			request.Configuration.ID,
			strings.ToLower(params.ByName(`ID`)),
		))
	}

	if _, err := uuid.FromString(request.Configuration.ID); err != nil {
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
	request.Configuration.ID = strings.ToLower(params.ByName(`ID`))

	if _, err := uuid.FromString(request.Configuration.ID); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	// request body may contain request flag overrides, API protocol v1
	// has no request body support
	if request.Version != msg.ProtocolOne {
		cReq := v2.NewConfigurationRequest()
		if err := decodeJSONBody(r, &cReq); err != nil {
			replyBadRequest(&w, &request, err)
			return
		}

		if err := resolveFlags(&cReq, &request); err != nil {
			replyBadRequest(&w, &request, err)
			return
		}
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
	request.Configuration.ID = strings.ToLower(params.ByName(`ID`))

	if _, err := uuid.FromString(request.Configuration.ID); err != nil {
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

// ConfigurationHistory accepts requests to retrieve the history of a
// configuration
func (x *Rest) ConfigurationHistory(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionConfiguration
	request.Action = msg.ActionHistory
	request.Configuration.ID = strings.ToLower(params.ByName(`ID`))

	if _, err := uuid.FromString(request.Configuration.ID); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`configuration_r`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// ConfigurationVersion accepts requests to retrieve a specific,
// possibly historic version of a Configuration
func (x *Rest) ConfigurationVersion(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)
	var err error

	request := msg.New(r, params)
	request.Section = msg.SectionConfiguration
	request.Action = msg.ActionVersion

	// configurationID is a mandatory URI path element, it may not
	// resolve to an empty string and must be a valid UUID
	request.Search.Configuration.ID = strings.TrimSpace(
		strings.ToLower(params.ByName(`ID`)),
	)
	if request.Search.Configuration.ID == `` {
		replyBadRequest(&w, &request, nil)
		return
	}
	if _, err = uuid.FromString(
		request.Search.Configuration.ID,
	); err != nil {
		replyBadRequest(&w, &request, err)
		return
	}

	// dataID is an optional URI path element. It may be empty, but if
	// it is not then it must be a valid UUID
	dataID := strings.TrimSpace(strings.TrimLeft(
		strings.ToLower(params.ByName(`DATA`)),
		`/`,
	))
	if dataID != `` {
		if _, err = uuid.FromString(dataID); err != nil {
			replyBadRequest(&w, &request, err)
			return
		}
		request.Search.Configuration.Data = []v2.Data{v2.Data{
			ID: dataID,
		}}
	} else {
		request.Search.Configuration.Data = []v2.Data{v2.Data{}}
	}

	// parse URL query parameters. The 'valid=...' query parameter is
	// optional, but if it is set then it must be a parsable RFC3339
	// timestamp
	if err = r.ParseForm(); err != nil {
		replyInternalError(&w, &request, err)
		return
	}
	if valid := r.Form.Get(`valid`); valid != `` {
		if request.Search.ValidAt, err = time.Parse(
			time.RFC3339Nano,
			valid,
		); err != nil {
			replyInternalError(&w, &request, err)
			return
		}
	}

	// while both dataID and valid are optional, one of them must be
	// provided
	if request.Search.ValidAt.IsZero() && dataID == `` {
		replyBadRequest(&w, &request, nil)
		return
	}

	if !x.isAuthorized(&request) {
		replyForbidden(&w, &request, nil)
		return
	}

	handler := x.handlerMap.Get(`configuration_r`)
	handler.Intake() <- request
	result := <-request.Reply
	respond(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
