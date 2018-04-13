/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import "github.com/julienschmidt/httprouter"

// setupRouter returns a configured httprouter
func (x *Rest) setupRouter() *httprouter.Router {
	router := httprouter.New()

	router.GET(`/api/v2/configuration/:lookup`, x.Verify(x.ConfigurationLookup))

	return router
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
