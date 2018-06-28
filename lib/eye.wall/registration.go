/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/mjolnir42/eye/lib/eye.wall"

import proto "github.com/mjolnir42/eye/lib/eye.proto"

// Register adds this eyewall cache to the list of active caches that
// must be invalidated by Eye
func (l *Lookup) Register() error {
	// apiVersion is not initialized, run a quick tasting
	if l.apiVersion == proto.ProtocolInvalid {
		l.taste(true)
	}

	switch l.apiVersion {
	case proto.ProtocolTwo:
		return l.v2Register()
	}

	return ErrProtocol
}

// Unregister removes the cache invalidation registration from Eye
func (l *Lookup) Unregister() error {
	// apiVersion is not initialized, can't possibly be registered
	if l.apiVersion == proto.ProtocolInvalid {
		l.taste(true)
	}

	switch l.apiVersion {
	case proto.ProtocolTwo:
		return l.v2Unregister()
	}

	return ErrProtocol
}

// LookupRegistrations returns the registrations for app
func (l *Lookup) LookupRegistrations(app string) (*proto.Result, error) {
	switch l.apiVersion {
	case proto.ProtocolTwo:
		return l.v2LookupRegistrations(app)
	}

	return nil, ErrProtocol
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
