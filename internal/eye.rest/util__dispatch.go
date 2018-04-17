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
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"

	"github.com/mjolnir42/soma/lib/proto"
)

func panicCatcher(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Printf("%s\n", debug.Stack())
		msg := fmt.Sprintf("PANIC! %s", r)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
}

// dispatchBadRequest returns a 400 error
func dispatchBadRequest(w *http.ResponseWriter, reason string) {
	if reason != `` {
		http.Error(*w, reason, http.StatusBadRequest)
		return
	}
	http.Error(*w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

// dispatchForbidden returns a 403 error
func dispatchForbidden(w *http.ResponseWriter, err error) {
	if err != nil {
		http.Error(*w, err.Error(), http.StatusForbidden)
		return
	}
	http.Error(*w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

// dispatchInternalError returns a 500 error
func dispatchInternalError(w *http.ResponseWriter, err error) {
	if err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(*w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// dispatchBadGateway returns a 502 error
func dispatchBadGateway(w *http.ResponseWriter, err error) {
	if err != nil {
		http.Error(*w, err.Error(), http.StatusBadGateway)
		return
	}
	http.Error(*w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
}

// dispatchGatewayTimeout returns a 504 error
func dispatchGatewayTimeout(w *http.ResponseWriter, err error) {
	if err != nil {
		http.Error(*w, err.Error(), http.StatusGatewayTimeout)
		return
	}
	http.Error(*w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
}

func decodeJSONBody(r *http.Request, s interface{}) (err error) {
	decoder := json.NewDecoder(r.Body)

	switch s.(type) {
	case *proto.PushNotification:
		c := s.(*proto.PushNotification)
		err = decoder.Decode(c)
	default:
		err = fmt.Errorf("decodeJSONBody: unhandled request type: %s", reflect.TypeOf(s))
	}
	return
}

func dispatchJSONReply(w *http.ResponseWriter, b *[]byte) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusOK)
	(*w).Write(*b)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
