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

	"github.com/julienschmidt/httprouter"
)

// Verify is a wrapper for CheckShutdown and BasicAuth checks
func (x *Rest) Verify(h httprouter.Handle) httprouter.Handle {
	return x.CheckShutdown(
		x.BasicAuth(
			func(w http.ResponseWriter, r *http.Request,
				ps httprouter.Params) {
				h(w, r, ps)
			},
		),
	)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
