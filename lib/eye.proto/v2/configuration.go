/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package v2 // import "github.com/mjolnir42/eye/lib/eye.proto/v2"

import "github.com/mjolnir42/eye/lib/eye.proto/v1"

// Configuration holds the monitoring profile definition for a check
// that has to be performed
type Configuration struct {
	ActivatedAt string `json:"activatedAt"`
	Data        []Data `json:"data"`
	HostID      uint64 `json:"hostID,string"`
	ID          string `json:"configurationID" valid:"uuidv4"`
	LookupID    string `json:"lookupID"`
	Metric      string `json:"metric"`
}

// Data contains a configuration
type Data struct {
	ID         string          `json:"dataID" valid:"uuidv4"`
	Info       MetaInformation `json:"information"`
	Interval   uint64          `json:"interval"`
	Monitoring string          `json:"monitoring"`
	Oncall     string          `json:"oncall"`
	Source     string          `json:"source"`
	Tags       []string        `json:"tags,omitempty"`
	Targethost string          `json:"targethost" valid:"host"`
	Team       string          `json:"string"`
	Thresholds []Threshold     `json:"thresholds"`
}

// Threshold contains the specification for a threshold of
// a Configuration
type Threshold struct {
	Predicate string `json:"predicate"`
	Level     uint16 `json:"level"`
	Value     int64  `json:"value"`
}

// MetaInformation contains registration metadata for the Configuration
type MetaInformation struct {
	ValidFrom       string   `json:"validFrom"`
	ValidUntil      string   `json:"validUntil"`
	ProvisionedAt   string   `json:"provisionedAt"`
	DeprovisionedAt string   `json:"deprovisionedAt"`
	Tasks           []string `json:"tasks"`
}

// NewConfigurationRequest returns a new request
func NewConfigurationRequest() Request {
	return Request{
		Flags:         &Flags{},
		Configuration: &Configuration{},
	}
}

// NewConfigurationResult returns a new result
func NewConfigurationResult() Result {
	return Result{
		Errors:         &[]string{},
		Configurations: &[]Configuration{},
	}
}

// InputSanatize ensures metadata fields are empty for c
func (c *Configuration) InputSanatize() {
	c.ActivatedAt = ``
	c.LookupID = ``
	for i, data := range c.Data {
		data.ID = ``
		data.Info = MetaInformation{}
		c.Data[i] = data
	}
}

// ConfigurationFromV1 converts configuration data between protocol
// versions v1 and v2
func ConfigurationFromV1(item *v1.ConfigurationItem) Configuration {
	cfg := Configuration{
		HostID: item.HostID,
		ID:     item.ConfigurationItemID,
		Metric: item.Metric,
	}
	data := Data{
		Interval:   item.Interval,
		Monitoring: item.Metadata.Monitoring,
		Oncall:     item.Oncall,
		Source:     item.Metadata.Source,
		Tags:       item.Tags,
		Targethost: item.Metadata.Targethost,
		Team:       item.Metadata.Team,
	}

	data.Thresholds = make([]Threshold, len(item.Thresholds))
	for i, thr := range item.Thresholds {
		data.Thresholds[i] = Threshold{
			Predicate: thr.Predicate,
			Level:     thr.Level,
			Value:     thr.Value,
		}
	}

	cfg.Data = []Data{
		data,
	}
	return cfg
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
