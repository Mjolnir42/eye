/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/mjolnir42/eye/lib/eye.wall"

import (
	"net/url"
	"strings"
)

// foldSlashes folds consecutive slashes in u.RequestURI
func foldSlashes(u *url.URL) {
	o := u.RequestURI()

	for u.Path = strings.Replace(
		u.RequestURI(), `//`, `/`, -1,
	); o != u.RequestURI(); u.Path = strings.Replace(
		u.RequestURI(), `//`, `/`, -1,
	) {
		o = u.RequestURI()
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
