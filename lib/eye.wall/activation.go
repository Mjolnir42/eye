/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/mjolnir42/eye/lib/eye.wall"

import proto "github.com/mjolnir42/eye/lib/eye.proto"

// Activate marks a profile as active if l detected an API version that
// supports profile Activation
func (l *Lookup) Activate(profileID string) error {
	// apiVersion is not initialized, run a quick tasting
	if l.apiVersion == proto.ProtocolInvalid {
		l.taste(true)
	}

	switch l.apiVersion {
	case proto.ProtocolTwo:
		return l.v2ActivateProfile(profileID)
	}

	return ErrProtocol
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
