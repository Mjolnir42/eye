/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/mjolnir42/eye/lib/eye.wall"

import proto "github.com/mjolnir42/eye/lib/eye.proto"

// ConfigurationShow ...
func (l *Lookup) ConfigurationShow(profileID string) (*proto.Result, error) {
	switch l.apiVersion {
	case proto.ProtocolOne:
		return l.v1ConfigurationShow(profileID)
	case proto.ProtocolTwo:
		return l.v2ConfigurationShow(profileID)
	}

	return nil, ErrProtocol
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
