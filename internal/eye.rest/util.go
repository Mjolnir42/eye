/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"

	somaproto "github.com/mjolnir42/soma/lib/proto"
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
	case *somaproto.PushNotification:
		c := s.(*somaproto.PushNotification)
		err = decoder.Decode(c)
	default:
		err = fmt.Errorf("decodeJSONBody: unhandled request type: %s", reflect.TypeOf(s))
	}
	return
}

// calculateLookupID returns the lookupID hash for a given (id,metric)
// tuple
func calculateLookupID(id uint64, metric string) string {
	asset := strconv.FormatUint(id, 10)
	hash := sha256.New()
	hash.Write([]byte(asset))
	hash.Write([]byte(metric))

	return hex.EncodeToString(hash.Sum(nil))
}

// getServiceAttributeValue returns the value of the requested service
// attribute or the empty string otherwise
func getServiceAttributeValue(details *somaproto.Deployment, attribute string) string {
	if details.Service == nil {
		return ``
	}
	if len(details.Service.Attributes) == 0 {
		return ``
	}
	for _, attr := range details.Service.Attributes {
		if attr.Name == attribute {
			return attr.Value
		}
	}
	return ``
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
