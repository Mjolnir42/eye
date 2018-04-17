/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import (
	"net/http"
)

// dispatchNoContent returns a 204 result
func dispatchNoContent(w *http.ResponseWriter) {
	http.Error(*w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
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

// dispatchGone returns a 410 error
func dispatchGone(w *http.ResponseWriter, err error) {
	if err != nil {
		http.Error(*w, err.Error(), http.StatusGone)
		return
	}
	http.Error(*w, http.StatusText(http.StatusGone), http.StatusGone)
}

// dispatchUnprocessableEntity returns a 422 error
func dispatchUnprocessableEntity(w *http.ResponseWriter, err error) {
	if err != nil {
		http.Error(*w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	http.Error(*w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
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

// dispatchJSONReply returns a 200 status JSON result
func dispatchJSONReply(w *http.ResponseWriter, b *[]byte) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusOK)
	(*w).Write(*b)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
