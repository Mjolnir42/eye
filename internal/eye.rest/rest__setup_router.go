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
	router.RedirectTrailingSlash = false

	router.DELETE(`/api/v1/item/:ID`, x.Verify(x.ConfigurationRemove))
	router.DELETE(`/api/v2/configuration/:ID`, x.Verify(x.ConfigurationRemove))
	router.DELETE(`/api/v2/registration/:ID`, x.Verify(x.RegistrationRemove))
	router.GET(`/api/v1/configuration/:hash`, x.Verify(x.LookupConfiguration))
	router.GET(`/api/v1/item/:ID`, x.Verify(x.ConfigurationShow))
	router.GET(`/api/v1/item/`, x.Verify(x.ConfigurationList))
	router.GET(`/api/v2/configuration/:ID/history/*DATA`, x.Verify(x.ConfigurationVersion))
	router.GET(`/api/v2/configuration/:ID/history`, x.Verify(x.ConfigurationHistory))
	router.GET(`/api/v2/configuration/:ID`, x.Verify(x.ConfigurationShow))
	router.GET(`/api/v2/configuration/`, x.Verify(x.ConfigurationList))
	router.GET(`/api/v2/lookup/configuration/:hash`, x.Verify(x.LookupConfiguration))
	router.GET(`/api/v2/lookup/registration/:application`, x.Verify(x.LookupRegistration))
	router.GET(`/api/v2/lookup/activation/`, x.Verify(x.LookupActivation))
	router.GET(`/api/v2/registration/:ID`, x.Verify(x.RegistrationShow))
	router.GET(`/api/v2/registration/`, x.Verify(x.RegistrationList))
	router.HEAD(`/api`, x.VersionInfo)
	router.PATCH(`/api/v2/configuration/:ID/active`, x.Verify(x.ConfigurationActivate))
	router.POST(`/api/v1/item/`, x.Verify(x.DeploymentProcess))
	router.POST(`/api/v1/notify/`, x.Verify(x.DeploymentNotification))
	router.POST(`/api/v1/notify`, x.Verify(x.DeploymentNotification))
	router.POST(`/api/v2/configuration/`, x.Verify(x.ConfigurationAdd))
	router.POST(`/api/v2/deployment/`, x.Verify(x.DeploymentProcess))
	router.POST(`/api/v2/deployment/notification`, x.Verify(x.DeploymentNotification))
	router.POST(`/api/v2/registration/`, x.Verify(x.RegistrationAdd))
	router.PUT(`/api/v1/item/:ID`, x.Verify(x.DeploymentProcess))
	router.PUT(`/api/v2/configuration/:ID`, x.Verify(x.ConfigurationUpdate))
	router.PUT(`/api/v2/registration/:ID`, x.Verify(x.RegistrationUpdate))

	return router
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
