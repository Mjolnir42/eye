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

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
