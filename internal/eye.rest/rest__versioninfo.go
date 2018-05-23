/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// EyeVersion can be set to version that Rest should report in
// VersionInfo requests
var EyeVersion = `undef`

// VersionInfo serves HEAD requests of the form /api?version=X where X
// is the inquired API version. For supported versions, the response
// will be 204/NoContent, 501/NotImplemented otherwise.
// Requests without ?version= query parameter are 400/BadRequest.
func (x *Rest) VersionInfo(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	var version int64
	var inq string
	var err error

	// set base info headers
	w.Header().Set(`X-Application-Info`, `EYE Monitoring Profile Server`)
	w.Header().Set(`X-Version`, EyeVersion)

	// parse URL query parameters
	if err = r.ParseForm(); err != nil {
		hardInternalError(&w)
		return
	}

	if inq = r.Form.Get(`version`); inq != `` {
		if version, err = strconv.ParseInt(inq, 10, 64); err != nil {
			hardInternalError(&w)
			return
		}
	} else {
		// version query parameter was not set
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch version {
	case 1:
		w.WriteHeader(http.StatusNoContent)
	case 2:
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
