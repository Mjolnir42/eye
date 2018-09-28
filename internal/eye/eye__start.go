/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package eye // import "github.com/solnx/eye/internal/eye"

import (
	mock "github.com/solnx/eye/internal/eye.mock"
)

// Start launches all application handlers
func (e *Eye) Start() {
	// supervisor must run first
	psv := mock.NewPermissiveSupervisor(e.conf)
	e.handlerMap.Add(`supervisor`, psv)
	e.handlerMap.Register(`supervisor`, e.dbConnection, e.exportLogger())
	e.handlerMap.Run(`supervisor`)

	// start regular handlers
	e.handlerMap.Add(`configuration_r`, newConfigurationRead(e.conf.Eye.QueueLen))
	e.handlerMap.Add(`configuration_w`, newConfigurationWrite(e.conf.Eye.QueueLen))
	e.handlerMap.Add(`deployment_w`, newDeploymentWrite(e.conf.Eye.QueueLen))
	e.handlerMap.Add(`lookup_r`, newLookupRead(e.conf.Eye.QueueLen))
	e.handlerMap.Add(`registration_r`, newRegistrationRead(e.conf.Eye.QueueLen))
	e.handlerMap.Add(`registration_w`, newRegistrationWrite(e.conf.Eye.QueueLen))

	for handler := range e.handlerMap.Range() {
		switch handler {
		case `supervisor`, `grimReaper`:
			// already running
			continue
		}
		e.handlerMap.Register(
			handler,
			e.dbConnection,
			e.exportLogger(),
		)
		// starts the handler in a goroutine
		e.handlerMap.Run(handler)
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
