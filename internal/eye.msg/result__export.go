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

	// v1 list results use status codes 200, 404, 500
	// v2 list results generate status codes 200, 403, 500, 501
	// this function maps 403+501 to 500 and generates 404 for empty 200
	// results
	if r.Error != nil {
		return ResultServerError, r.Error.Error(), list
	}

	// v1 returned 404 on empty list results
	if len(r.Configuration) == 0 {
		return ResultNotFound, http.StatusText(ResultNotFound), list
	}

	list.ConfigurationItemIDList = make([]string, len(r.Configuration))
	for i, id := range r.Configuration {
		list.ConfigurationItemIDList[i] = id.ID
	}
	return ResultOK, ``, list
}

// ExportV1ConfigurationShow generates a protocol version 1 show result
func (r *Result) ExportV1ConfigurationShow() (uint16, string, v1.ConfigurationData) {
	cfg := v1.ConfigurationData{}

	// v1 show results use status codes 200, 400, 404, 500
	// v2 show results generate status codes 200, 400, 403, 404, 500, 501
	switch r.Code {
	case ResultBadRequest, ResultNotFound, ResultServerError:
		return r.Code, r.Error.Error(), cfg
	case ResultForbidden, ResultNotImplemented:
		return ResultServerError, r.Error.Error(), cfg
	}

	if len(r.Configuration) != 1 {
		// internal result has been generated incorrectly
		return ResultServerError, http.StatusText(ResultServerError), cfg
	}

	res := r.Configuration[0]
	data := res.Data[0]

	cfg.Configurations = make([]v1.ConfigurationItem, 1)
	item := v1.ConfigurationItem{
		ConfigurationItemID: res.ID,
		Metric:              res.Metric,
		HostID:              res.HostID,
		Tags:                data.Tags,
		Oncall:              data.Oncall,
		Interval:            data.Interval,
		Metadata: v1.ConfigurationMetaData{
			Monitoring: data.Monitoring,
			Team:       data.Team,
			Source:     data.Source,
			Targethost: data.Targethost,
		},
		Thresholds: []v1.ConfigurationThreshold{},
	}

	for _, thr := range data.Thresholds {
		item.Thresholds = append(item.Thresholds, v1.ConfigurationThreshold{
			Predicate: thr.Predicate,
			Level:     thr.Level,
			Value:     thr.Value,
		})
	}
	cfg.Configurations[0] = item
	return ResultOK, ``, cfg
}

// ExportV1LookupCfg generates a protocol version 1 lookup result
func (r *Result) ExportV1LookupCfg() (uint16, string, v1.ConfigurationData) {
	cfg := v1.ConfigurationData{}

	// v1 lookup results use status codes 200, 400,      404, 500
	// v2 lookup results use status codes 200, 400, 403, 404, 500, 501
	switch r.Code {
	case ResultBadRequest, ResultNotFound, ResultServerError:
		return r.Code, r.Error.Error(), cfg
	case ResultForbidden, ResultNotImplemented:
		return ResultServerError, r.Error.Error(), cfg
	}

	cfg.Configurations = make(
		[]v1.ConfigurationItem,
		len(r.Configuration),
	)

	for idx, res := range r.Configuration {
		data := res.Data[0]

		item := v1.ConfigurationItem{
			ConfigurationItemID: res.ID,
			Metric:              res.Metric,
			HostID:              res.HostID,
			Tags:                data.Tags,
			Oncall:              data.Oncall,
			Interval:            data.Interval,
			Metadata: v1.ConfigurationMetaData{
				Monitoring: data.Monitoring,
				Team:       data.Team,
				Source:     data.Source,
				Targethost: data.Targethost,
			},
			Thresholds: []v1.ConfigurationThreshold{},
		}
		for _, thr := range data.Thresholds {
			item.Thresholds = append(item.Thresholds, v1.ConfigurationThreshold{
				Predicate: thr.Predicate,
				Level:     thr.Level,
				Value:     thr.Value,
			})
		}

		cfg.Configurations[idx] = item
	}

	return ResultOK, ``, cfg
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
