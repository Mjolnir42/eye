/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/mjolnir42/eye/lib/eye.proto"
	somaproto "github.com/mjolnir42/soma/lib/proto"
)

func panicCatcher(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Printf("%s\n", debug.Stack())
		msg := fmt.Sprintf("PANIC! %s", r)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
}

func decodeJSONBody(r *http.Request, s interface{}) (err error) {
	decoder := json.NewDecoder(r.Body)

	switch s.(type) {
	case *somaproto.PushNotification:
		c := s.(*somaproto.PushNotification)
		err = decoder.Decode(c)
	default:
		err = fmt.Errorf("decodeJSONBody: unhandled request type: %s", reflect.TypeOf(s))
	}
	return
}

// processDeploymentDetails creates an eye protocol configuration from
// SOMA deployment details
func processDeploymentDetails(details *somaproto.Deployment) (string, *proto.Configuration, error) {
	lookupID := calculateLookupID(details.Node.AssetID, details.Metric.Path)

	config := &proto.Configuration{
		ConfigurationID: details.CheckInstance.InstanceID,
		Metric:          details.Metric.Path,
		Interval:        details.CheckConfig.Interval,
		//HostID:   strconv.FormatUint(details.Node.AssetID, 10),
		HostID: details.Node.AssetID,
		Metadata: proto.MetaData{
			Monitoring: details.Monitoring.Name,
			Team:       details.Team.Name,
		},
		Thresholds: []proto.Threshold{},
	}

	// append filesystem to disk metrics
	switch config.Metric {
	case
		`disk.write.per.second`,
		`disk.read.per.second`,
		`disk.free`,
		`disk.usage.percent`:
		mountpoint := getServiceAttributeValue(details, `filesystem`)
		if mountpoint == `` {
			return ``, nil, fmt.Errorf("Disk metric %s is missing filesystem service attribute", config.Metric)
		}

		// update metric path and recalculate updated lookupID
		config.Metric = fmt.Sprintf("%s:%s", config.Metric, mountpoint)
		lookupID = calculateLookupID(details.Node.AssetID, details.Metric.Path)
	}

	// set oncall duty
	if details.Oncall != nil && details.Oncall.ID != `` {
		config.Oncall = fmt.Sprintf("%s (%s)", details.Oncall.Name, details.Oncall.Number)
	}

	config.Metadata.Targethost = getTargetHost(details)

	// construct item.Metadata.Source
	if details.Service != nil && details.Service.Name != `` {
		config.Metadata.Source = fmt.Sprintf("%s, %s", details.Service.Name, details.CheckConfig.Name)
	} else {
		config.Metadata.Source = fmt.Sprintf("System (%s), %s", details.Node.Name, details.CheckConfig.Name)
	}

	// slurp all thresholds
	for _, thr := range details.CheckConfig.Thresholds {
		t := proto.Threshold{
			Predicate: thr.Predicate.Symbol,
			Level:     thr.Level.Numeric,
			Value:     thr.Value,
		}
		config.Thresholds = append(config.Thresholds, t)
	}

	govalidator.SetFieldsRequiredByDefault(true)
	if ok, err := govalidator.ValidateStruct(config); !ok {
		return ``, nil, err
	}
	return lookupID, config, nil
}

// calculateLookupID returns the lookupID hash for a given (id,metric)
// tuple
func calculateLookupID(id uint64, metric string) string {
	asset := strconv.FormatUint(id, 10)
	hash := sha256.New()
	hash.Write([]byte(asset))
	hash.Write([]byte(metric))

	return hex.EncodeToString(hash.Sum(nil))
}

// getServiceAttributeValue returns the value of the requested service
// attribute or the empty string otherwise
func getServiceAttributeValue(details *somaproto.Deployment, attribute string) string {
	if details.Service == nil {
		return ``
	}
	if len(details.Service.Attributes) == 0 {
		return ``
	}
	for _, attr := range details.Service.Attributes {
		if attr.Name == attribute {
			return attr.Value
		}
	}
	return ``
}

// getTargetHost returns the hostname of the deployment target
func getTargetHost(details *somaproto.Deployment) string {
	var fqdn, dnsZone string

	// details.Properties contains only system properties which are
	// guaranteed to be unique by the SOMA data model
	if details.Properties != nil {
		for _, prop := range *details.Properties {
			switch prop.Name {
			case `fqdn`:
				fqdn = prop.Value
			case `dns_zone`:
				dnsZone = prop.Value
			}
		}
	}

	switch {
	// specified fqdn has the highest priority
	case fqdn != ``:
		return fqdn

		// trailing dot prevents appending the configured zone
	case strings.HasSuffix(details.Node.Name, `.`):
		return details.Node.Name

		// configured zone is appended to the hostname
	case dnsZone != ``:
		return fmt.Sprintf("%s.%s", details.Node.Name, dnsZone)

		// no better data available
	default:
		return details.Node.Name
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
