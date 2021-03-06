/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto // import "github.com/solnx/eye/lib/eye.proto"

import (
	"github.com/solnx/eye/lib/eye.proto/v1"
	"github.com/solnx/eye/lib/eye.proto/v2"
)

// Result wraps Results for multiple versions
type Result struct {
	APIVersion int                   `json:"apiVersion"`
	V1Data     *v1.ConfigurationData `json:"v1Data,omitempty"`
	V2Result   *v2.Result            `json:"v2Result,omitempty"`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
