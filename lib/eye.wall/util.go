/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/solnx/eye/lib/eye.wall"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/solnx/eye/lib/eye.proto/v1"
	"github.com/solnx/eye/lib/eye.proto/v2"
)

// foldSlashes folds consecutive slashes in u.RequestURI
func foldSlashes(u *url.URL) {
	o := u.RawPath
	for u.RawPath = strings.Replace(
		u.RawPath, `//`, `/`, -1,
	); o != u.RawPath; u.RawPath = strings.Replace(
		u.RawPath, `//`, `/`, -1,
	) {
		o = u.RawPath
	}
}

// v1ConfigurationData returns a deserialized v1.ConfigurationData from
// a response body
func v1ConfigurationData(body []byte) (data *v1.ConfigurationData, err error) {
	if err = json.Unmarshal(body, data); err != nil {
		return
	}
	return
}

// v2Result returns a deserialized v2.Result from a response body
func v2Result(body []byte) (result *v2.Result, err error) {
	result = &v2.Result{}
	if err = json.Unmarshal(body, result); err != nil {
		return
	}

	// Protocol2 always responds 200 as HTTP code if the request could
	// be routed to the application
	switch result.StatusCode {
	case http.StatusOK:

		// success
	case http.StatusNotFound:
		result = nil
		err = ErrUnconfigured
	default:
		// there was some error
		err = fmt.Errorf("eye(%s|%s) %d/%s: %v",
			result.Section,
			result.Action,
			result.StatusCode,
			result.StatusText,
			result.Errors,
		)
		result = nil
	}
	return
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
