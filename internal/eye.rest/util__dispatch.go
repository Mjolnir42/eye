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
	"log"
	"net/http"
	"runtime/debug"
)

func dispatchForbidden(w *http.ResponseWriter, err error) {
	if err != nil {
		http.Error(*w, err.Error(), http.StatusForbidden)
		return
	}
	http.Error(*w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

func dispatchBadRequest(w *http.ResponseWriter, reason string) {
	if reason != `` {
		http.Error(*w, reason, http.StatusBadRequest)
		return
	}
	http.Error(*w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func panicCatcher(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Printf("%s\n", debug.Stack())
		msg := fmt.Sprintf("PANIC! %s", r)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
}

func dispatchJSONReply(w *http.ResponseWriter, b *[]byte) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusOK)
	(*w).Write(*b)
}

func dispatchInternalError(w *http.ResponseWriter, err error) {
	if err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(*w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
