/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/solnx/eye/internal/eye.rest"

import (
	"encoding/json"
	"net/http"

	msg "github.com/solnx/eye/internal/eye.msg"
	"github.com/solnx/eye/lib/eye.proto/v2"
)

// respond is the output function for all requests
func (x *Rest) respond(w *http.ResponseWriter, r *msg.Result) {
	switch r.Version {
	case msg.ProtocolInvalid:
		panic(`API Protocol 0 is not valid`)
	case msg.ProtocolOne:
		x.respondV1(w, r)
	case msg.ProtocolTwo:
		x.respondV2(w, r)
	default:
		panic(`API Protocol unknown`)
	}
}

// respondV1 is the output function emitting API version 1 results
func (x *Rest) respondV1(w *http.ResponseWriter, r *msg.Result) {
	var bjson []byte
	var err error
	// not available via v1
	r.Flags.CacheInvalidation = false
	r.Flags.AlarmClearing = false

	switch r.Section {
	case msg.SectionRegistration:
		panic(`API Protocol 1 does not have registrations`)

	case msg.SectionLookup:
		switch r.Action {
		case msg.ActionConfiguration:
			code, errstr, data := r.ExportV1LookupCfg()
			if bjson, err = json.Marshal(&data); err != nil {
				hardInternalError(w)
				return
			}
			sendV1Result(w, code, errstr, &bjson)

		default:
			hardInternalError(w)
			return
		}

	case msg.SectionConfiguration:
		switch r.Action {
		case msg.ActionList:
			code, errstr, list := r.ExportV1ConfigurationList()
			if bjson, err = json.Marshal(&list); err != nil {
				hardInternalError(w)
				return
			}
			sendV1Result(w, code, errstr, &bjson)
		case msg.ActionShow:
			code, errstr, data := r.ExportV1ConfigurationShow()
			if bjson, err = json.Marshal(&data); err != nil {
				hardInternalError(w)
				return
			}
			sendV1Result(w, code, errstr, &bjson)
		case msg.ActionAdd, msg.ActionUpdate:
			switch r.Code {
			case msg.ResultUnprocessable:
				sendV1Result(w, r.Code, r.Error.Error(), nil)
			}
			// ResultUnprocessable handling is the only difference
			// between Add|Update and Remove
			fallthrough
		case msg.ActionRemove:
			switch r.Code {
			case msg.ResultServerError, msg.ResultBadRequest:
				sendV1Result(w, r.Code, r.Error.Error(), nil)
			case msg.ResultForbidden:
				// v1 API has no 403/Forbidden
				sendV1Result(w, msg.ResultBadRequest, r.Error.Error(), nil)
			case msg.ResultNotImplemented:
				// v1 API has no 501/NotImplemented
				sendV1Result(w, msg.ResultServerError, r.Error.Error(), nil)
			case msg.ResultOK:
				// v1 API uses 204/NoContent
				sendV1Result(w, msg.ResultNoContent, ``, nil)
			case msg.ResultUnprocessable:
				// not an error case on fallthrough: ignore

				if r.Action == msg.ActionRemove {
					// error case for Remove
					hardInternalError(w)
					return
				}
			default:
				hardInternalError(w)
				return
			}
		default:
			hardInternalError(w)
			return
		}

	case msg.SectionDeployment:
		// v1 Deployment API uses: 204, 400,      410, 412, 422, 500
		// v2 Deployment API uses:      400, 403, 410,      422, 500, 501, 502, 504
		// only failed requests return in SectionDeployment before being
		// mapped to SectionConfiguration
		switch r.Code {
		case msg.ResultForbidden, msg.ResultNotImplemented:
			// v1 API has no 403/Forbidden or 501/NotImplemented
			sendV1Result(w, msg.ResultServerError, r.Error.Error(), nil)

		case msg.ResultBadGateway, msg.ResultGatewayTimeout:
			// v1 API uses 412/PreconditionFailed for connectivity
			// errors to SOMA which use 502/504 for the v2 API
			sendV1Result(w, http.StatusPreconditionFailed, r.Error.Error(), nil)

		case msg.ResultBadRequest, msg.ResultGone, msg.ResultUnprocessable, msg.ResultServerError:
			// directly mapped v1:v2 result codes
			sendV1Result(w, r.Code, r.Error.Error(), nil)

		default:
			// invalid unmapped result
			hardInternalError(w)
			return
		}

	default:
		hardInternalError(w)
		return
	}

	if r.Flags.SendDeploymentFeedback {
		go x.somaStatusUpdate(r)
	}
}

// respondV2 is the output function emitting API version 2 results
func (x *Rest) respondV2(w *http.ResponseWriter, r *msg.Result) {
	var (
		bjson    []byte
		err      error
		protoRes v2.Result
	)

	// create external protocol result
	switch r.Section {
	case msg.SectionConfiguration:
		protoRes = v2.NewConfigurationResult()
	case msg.SectionDeployment:
		protoRes = v2.NewConfigurationResult()
	case msg.SectionLookup:
		switch r.Action {
		case msg.ActionConfiguration, msg.ActionActivation, msg.ActionPending:
			protoRes = v2.NewConfigurationResult()
		case msg.ActionRegistration:
			protoRes = v2.NewRegistrationResult()
		}
	case msg.SectionRegistration:
		protoRes = v2.NewRegistrationResult()

		// update cache registry
		//x.eyewallCacheUnregister(r)
		//x.eyewallCacheRegister(r)
	}
	// record what was performed
	protoRes.Section = r.Section
	protoRes.Action = r.Action

	// internal result contains an error, copy over into protocol result
	if r.Error != nil {
		x.appLog.Tracef("Section=%s Action=%s Error=%s", protoRes.Section, protoRes.Action, r.Error.Error())
		*protoRes.Errors = append(*protoRes.Errors, r.Error.Error())
	}

	// copy internal result data into protocol result
	switch r.Section {
	case msg.SectionConfiguration:
		*protoRes.Configurations = append(*protoRes.Configurations, r.Configuration...)
	case msg.SectionDeployment:
		*protoRes.Configurations = append(*protoRes.Configurations, r.Configuration...)
	case msg.SectionLookup:
		switch r.Action {
		case msg.ActionConfiguration, msg.ActionActivation, msg.ActionPending:
			*protoRes.Configurations = append(*protoRes.Configurations, r.Configuration...)
		case msg.ActionRegistration:
			*protoRes.Registrations = append(*protoRes.Registrations, r.Registration...)
		}
	case msg.SectionRegistration:
		*protoRes.Registrations = append(*protoRes.Registrations, r.Registration...)
	}

	// trigger omitempty JSON encoding conditions if applicable
	if protoRes.Configurations != nil && len(*protoRes.Configurations) == 0 {
		*protoRes.Configurations = nil
	}
	if protoRes.Registrations != nil && len(*protoRes.Registrations) == 0 {
		*protoRes.Registrations = nil
	}

	// set protocol result status
	protoRes.SetStatus(r.Code)

	switch {
	// no results are exported on error to avoid accidental data leaks
	// no cache invalidation for failed requests
	// no alarm clearing for failed requests
	case r.Code >= 400:
		protoRes.Configurations = nil
		protoRes.Registrations = nil
		r.Flags.CacheInvalidation = false
		r.Flags.AlarmClearing = false
	}

	// perform cache invalidation
	x.eyewallCacheInvalidate(r)

	// send notification alarm event
	if r.Flags.AlarmClearing {
		go x.alarmSend(r)
	}
	// send deployment feedback to SOMA
	if r.Flags.SendDeploymentFeedback {
		go x.somaStatusUpdate(r)
	}

	if bjson, err = json.Marshal(&protoRes); err != nil {
		x.appLog.Errorf("Section=%s Action=%s Error=%s", "Respond", "V2", err.Error())
		hardInternalError(w)
		return
	}

	sendJSONReply(w, &bjson)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
