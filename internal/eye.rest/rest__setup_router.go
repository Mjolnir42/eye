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

	router.GET(`/api/v1/configuration/:hash`, x.Verify(x.LookupConfiguration))
	router.GET(`/api/v2/lookup/configuration/:hash`, x.Verify(x.LookupConfiguration))
	router.POST(`/api/v1/notify/`, x.Verify(x.DeploymentNotification))
	router.POST(`/api/v1/notify`, x.Verify(x.DeploymentNotification))
	router.POST(`/api/v2/deployment/notification`, x.Verify(x.DeploymentNotification))
	router.GET(`/api/v1/item/:ID`, x.Verify(x.ConfigurationShow))
	router.GET(`/api/v2/configuration/:ID`, x.Verify(x.ConfigurationShow))

	return router
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
