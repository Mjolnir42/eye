/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import (
	"encoding/json"
	"net/http"

	msg "github.com/mjolnir42/eye/internal/eye.msg"
	proto "github.com/mjolnir42/eye/lib/eye.proto"
)

// sendMsgResult is the output function for all requests that did not
// fail input validation and got processes by the application.
func sendMsgResult(w *http.ResponseWriter, r *msg.Result) {
	var (
		bjson                []byte
		err                  error
		feedback, clearAlarm bool
		feedbackType         string
	)
	result := proto.NewConfigurationResult()

	// internal result contains an error, copy over into protocol result
	if r.Error != nil {
		*result.Errors = append(*result.Errors, r.Error.Error())
		feedbackType = `failed`
	}

	// copy internal result data into protocol result
	switch r.Section {
	case msg.SectionLookup:
		*result.Configurations = append(*result.Configurations, r.Configuration...)

	case msg.SectionDeployment:
		// only errors return with r.Section == msg.SectionDeployment
		*result.Configurations = nil
		feedback = true
		feedbackType = `failed`

	case msg.SectionConfiguration:
		switch r.ConfigurationTask {
		// configuration action originated from a push notification deployment
		case msg.TaskDelete:
			clearAlarm = true
			fallthrough
		case msg.TaskRollout, msg.TaskPending, msg.TaskDeprovision:
			feedback = true
		}

	default:
		dispatchInternalError(w, nil)
		return
	}

	// set protocol result status
	switch r.Code {
	case msg.ResultOK:
		result.SetStatus(r.Code)
		feedbackType = `success`
	case msg.ResultServerError, msg.ResultNotImplemented:
		result.SetStatus(r.Code)

	default:
		dispatchInternalError(w, nil)
		return
	}

	if feedback {
		go sendSomaFeedback(r.FeedbackURL, feedbackType)
	}

	if clearAlarm && !r.HasFailed() && r.Action == msg.ActionRemove {
		go clearCamsAlarm(r)
	}

	if bjson, err = json.Marshal(&result); err != nil {
		dispatchInternalError(w, nil)
		return
	}

	dispatchJSONReply(w, &bjson)
	return
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
