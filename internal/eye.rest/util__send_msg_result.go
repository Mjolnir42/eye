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
		bjson  []byte
		err    error
		result proto.Result
	)

	switch r.Section {
	case msg.SectionConfiguration:
		result = proto.NewConfigurationResult()
		*result.Configurations = append(*result.Configurations, r.Configuration...)

	default:
		dispatchInternalError(w, nil)
		return
	}

	switch r.Code {
	case msg.ResultOK:
		result.OK()

	default:
		dispatchInternalError(w, nil)
		return
	}

	if bjson, err = json.Marshal(&result); err != nil {
		dispatchInternalError(w, nil)
		return
	}

	dispatchJSONReply(w, &bjson)
	return
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
