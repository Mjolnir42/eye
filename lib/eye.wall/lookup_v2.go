/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/mjolnir42/eye/lib/eye.wall"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty"
	"github.com/mjolnir42/eye/lib/eye.proto/v2"
)

// v2LookupEye queries the Eye monitoring profile server
func (l *Lookup) v2LookupEye(lookID string) (*v2.Result, error) {
	var err error
	var resp *resty.Response

	if resp, err = l.client.R().
		SetPathParams(map[string]string{
			`lookID`: lookID,
		}).Get(
		l.eyeLookupURL.String(),
	); err != nil {
		return nil, fmt.Errorf("eyewall.Lookup: %s", err.Error())
	}

	// Protocol2 always responds 200 as HTTP code if the request could
	// be routed to the application
	switch resp.StatusCode() {
	case http.StatusOK:
	default:
		return nil, fmt.Errorf("eyewall.Lookup: %s", resp.String())
	}

	result := &v2.Result{}
	if err = json.Unmarshal(resp.Body(), result); err != nil {
		return nil, fmt.Errorf("eyewall.Lookup: %s", err.Error())
	}

	switch result.StatusCode {
	case http.StatusOK:
		// success
	case http.StatusNotFound:
		// no profiles for lookID
		l.setUnconfigured(lookID)
		return nil, ErrUnconfigured
	default:
		// there was some error
		return nil, fmt.Errorf("eyewall.Lookup: eye(%s|%s) %d/%s: %v",
			result.Section,
			result.Action,
			result.StatusCode,
			result.StatusText,
			result.Errors,
		)
	}
	return result, nil
}

// v2Process converts t into Threshold and stores it in the
// local cache if available
func (l *Lookup) v2Process(lookID string, pr *v2.Result) (map[string]Threshold, error) {
	if pr.Configurations == nil {
		return nil, fmt.Errorf(`eyewall.Lookup: v2Process received pr.Configurations == nil`)
	}
	if len(*pr.Configurations) == 0 {
		l.setUnconfigured(lookID)
		return nil, ErrUnconfigured
	}

	res := make(map[string]Threshold)
	for _, i := range *pr.Configurations {
		t := Threshold{
			ID:             i.ID,
			Metric:         i.Metric,
			HostID:         i.HostID,
			Oncall:         i.Data[0].Oncall,
			Interval:       i.Data[0].Interval,
			MetaMonitoring: i.Data[0].Monitoring,
			MetaTeam:       i.Data[0].Team,
			MetaSource:     i.Data[0].Source,
			MetaTargethost: i.Data[0].Targethost,
		}
		l.v2UpdateCachedActivation(i.ID, i.ActivatedAt)

		t.Thresholds = make(map[string]int64)
		for _, tl := range i.Data[0].Thresholds {
			lvl := strconv.FormatUint(uint64(tl.Level), 10)
			t.Predicate = tl.Predicate
			t.Thresholds[lvl] = tl.Value
		}

		l.storeThreshold(lookID, &t)
		res[t.ID] = t
	}
	return res, nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
