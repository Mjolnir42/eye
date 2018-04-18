/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto // import "github.com/mjolnir42/eye/lib/eye.proto"

// Configuration holds the monitoring profile definition for a check
// that has to be performed
type Configuration struct {
	ID         string      `json:"configurationID" valid:"uuidv4"`
	Metric     string      `json:"metric"`
	HostID     uint64      `json:"hostID,string"`
	Tags       []string    `json:"tags,omitempty"`
	Oncall     string      `json:"oncall"`
	Interval   uint64      `json:"interval"`
	Metadata   MetaData    `json:"metadata"`
	Thresholds []Threshold `json:"thresholds"`
}

// MetaData contains the metadata for a Configuration
type MetaData struct {
	Monitoring string `json:"monitoring"`
	Team       string `json:"string"`
	Source     string `json:"source"`
	Targethost string `json:"targethost" valid:"host"`
}

// Threshold contains the specification for a threshold of
// a Configuration
type Threshold struct {
	Predicate string `json:"predicate"`
	Level     uint16 `json:"level"`
	Value     int64  `json:"value"`
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

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
