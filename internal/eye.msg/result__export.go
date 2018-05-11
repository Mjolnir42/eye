/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package msg // import "github.com/mjolnir42/eye/internal/eye.msg"

import (
	"net/http"

	"github.com/mjolnir42/eye/lib/eye.proto/v1"
)

// ExportV1ConfigurationList generates a protocol version 1 list result
func (r *Result) ExportV1ConfigurationList() (uint16, string, v1.ConfigurationList) {
	list := v1.ConfigurationList{}
	if r.Error != nil {
		return r.Code, r.Error.Error(), list
	}

	// v1 returned 404 on empty list results
	if len(r.Configuration) == 0 {
		return ResultNotFound, http.StatusText(ResultNotFound), list
	}

	list.ConfigurationItemIDList = make([]string, len(r.Configuration))
	for i, id := range r.Configuration {
		list.ConfigurationItemIDList[i] = id
	}
	return r.Code, ``, list
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
