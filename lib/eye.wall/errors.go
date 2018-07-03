/*-
 * Copyright © 2016,2017, Jörg Pernfuß <code.jpe@gmail.com>
 * Copyright © 2016, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/mjolnir42/eye/lib/eye.wall"

import (
	"errors"
)

var (
	// ErrNotFound is returned when the cache contains no matching data
	ErrNotFound = errors.New("eyewall.Lookup: not found")
	// ErrUnconfigured is returned when the cache contains a negative
	// caching entry or Eye returns the absence of a profile to look up
	ErrUnconfigured = errors.New("eyewall.Lookup: unconfigured")
	// ErrUnavailable is returned when the cache does not contain the
	// requested record and Eye can not be queried
	ErrUnavailable = errors.New(`eyewall.Lookup: profile server unavailable`)
	// ErrNoCache is returned if the application does not start the
	// local cache, but a request against the local cache is issued
	ErrNoCache = errors.New(`eyewall.Lookup: local cache not configured`)
	// ErrProtocol is returned if the application attempted an action
	// that is not supported by the used protocol version
	ErrProtocol = errors.New(`eyewall.Lookup: request unsupported by protocol version`)
)

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
