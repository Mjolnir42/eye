/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/mjolnir42/eye/lib/eye.wall"

import (
	"fmt"
	"time"

	"github.com/go-resty/resty"
)

// v2ActivateProfile implements the activation of profileID for API
// version 2
func (l *Lookup) v2ActivateProfile(profileID string) error {
	// check if the profile is already known activated
	if val, err := l.redis.HGet(
		`activation`,
		profileID,
	).Result(); err != nil {
		if l.log != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
		}
		return err
	} else if val != `never` {
		// profile is already marked activated inside the Cache
		return nil
	}

	// activate profile in Eye
	var err error
	var resp *resty.Response

	if resp, err = l.client.R().
		SetPathParams(map[string]string{
			`profileID`: profileID,
		}).Patch(
		l.eyeActiveURL.String(),
	); err != nil {
		return fmt.Errorf("eyewall/cache: %s", err.Error())
	}

	if _, err = v2Result(resp.Body()); err != nil {
		return err
	}

	// update activation in cache
	return l.v2UpdateCachedActivation(
		profileID,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
}

// v2UpdateCachedActivation writes profile activation information into
// redis. It is intended to be used with information loaded from eye and
// updates the cache unconditionally.
func (l *Lookup) v2UpdateCachedActivation(profileID, ts string) error {
	if _, err := l.redis.HSet(
		`activation`,
		profileID,
		ts,
	).Result(); err != nil {
		if l.log != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
		}
		return err
	}
	return nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
