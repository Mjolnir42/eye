/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto // import "github.com/mjolnir42/eye/lib/eye.proto"

// Request represents a v2 API request
type Request struct {
	Flags         *Flags         `json:"flags,omitempty"`
	Configuration *Configuration `json:"configuration,omitempty"`
	Registration  *Registration  `json:"registration,omitempty"`
}

// Flags contains the flags that a v2 API request can contain
type Flags struct {
	AlarmClearing          string `json:"alarm.clearing"`
	CacheInvalidation      string `json:"enable.cache.invalidation"`
	SendDeploymentFeedback string `json:"send.deployment.feedback"`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
