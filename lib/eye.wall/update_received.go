/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/solnx/eye/lib/eye.wall"

// UpdateReceived increments the counter of received metrics
func (l *Lookup) UpdateReceived() {
	l.pipe.Incr(`received_metrics`)
}

// resetReceived is called during startup to reset the number of
// received metrics set within the cache
func (l *Lookup) resetReceived() error {
	if _, err := l.redis.Set(
		`received_metrics`,
		`0`,
		0,
	).Result(); err != nil {
		if l.log != nil {
			l.log.Errorf("eyewall/resetReceived: %s", err.Error())
		}
		return err
	}
	return nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
