/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/solnx/eye/internal/eye.rest"

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/mjolnir42/soma/lib/proto"
	msg "github.com/solnx/eye/internal/eye.msg"
	eyeproto "github.com/solnx/eye/lib/eye.proto"
	"github.com/solnx/eye/lib/eye.proto/v2"
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
	case *proto.PushNotification:
		c := s.(*proto.PushNotification)
		err = decoder.Decode(c)
	case *v2.Request:
		c := s.(*v2.Request)
		err = decoder.Decode(c)
	default:
		err = fmt.Errorf("decodeJSONBody: unhandled request type: %s", reflect.TypeOf(s))
	}
	return
}

// processDeploymentDetails creates an eye protocol configuration from
// SOMA deployment details
func processDeploymentDetails(details *proto.Deployment) (string, v2.Configuration, error) {
	lookupID := calculateLookupID(details.Node.Name, details.Metric.Path)

	config := v2.Configuration{
		Hostname: details.Node.Name,
		HostID:   details.Node.AssetID,
		ID:       details.CheckInstance.InstanceID,
		Metric:   details.Metric.Path,
	}
	data := v2.Data{
		Interval:   details.CheckConfig.Interval,
		Monitoring: details.Monitoring.Name,
		Team:       details.Team.Name,
		Thresholds: []v2.Threshold{},
	}

	config.LookupID = lookupID

	// set oncall duty
	if details.Oncall != nil && details.Oncall.ID != `` {
		data.Oncall = fmt.Sprintf("%s (%s)", details.Oncall.Name, details.Oncall.Number)
	}

	data.Targethost = getTargetHost(details)

	// construct item.Metadata.Source
	if details.Service != nil && details.Service.Name != `` {
		data.Source = fmt.Sprintf("%s, %s", details.Service.Name, details.CheckConfig.Name)
	} else {
		data.Source = fmt.Sprintf("System (%s), %s", details.Node.Name, details.CheckConfig.Name)
	}

	// slurp all thresholds
	for _, thr := range details.CheckConfig.Thresholds {
		t := v2.Threshold{
			Predicate: thr.Predicate.Symbol,
			Level:     thr.Level.Numeric,
			Value:     thr.Value,
		}
		data.Thresholds = append(data.Thresholds, t)
	}
	config.Data = []v2.Data{data}

	govalidator.SetFieldsRequiredByDefault(false)
	if ok, err := govalidator.ValidateStruct(config); !ok {
		return ``, v2.Configuration{}, err
	}
	return lookupID, config, nil
}

// calculateLookupID returns the lookupID hash for a given (id,metric)
// tuple
func calculateLookupID(host, metric string) string {
	hash := sha256.New()
	hash.Write([]byte(host))
	hash.Write([]byte(metric))

	return hex.EncodeToString(hash.Sum(nil))
}

// getServiceAttributeValue returns the value of the requested service
// attribute or the empty string otherwise
func getServiceAttributeValue(details *proto.Deployment, attribute string) string {
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
func getTargetHost(details *proto.Deployment) string {
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

// resolveFlags sets the request flags of rqInternal based on the user
// input in rqProtocol as well as the request type
func resolveFlags(rqProtocol *v2.Request, rqInternal *msg.Request) error {
	switch rqInternal.Section {
	case msg.SectionConfiguration:
		switch rqInternal.Action {
		case msg.ActionRemove:
			if val, err := strconv.ParseBool(rqProtocol.Flags.AlarmClearing); err != nil {
				// disable by default
				rqInternal.Flags.AlarmClearing = false
			} else if val {
				// explicit enable
				rqInternal.Flags.AlarmClearing = true
			} else {
				rqInternal.Flags.AlarmClearing = false
			}
			fallthrough

		case msg.ActionAdd, msg.ActionUpdate:
			if val, err := strconv.ParseBool(rqProtocol.Flags.ResetActivation); err != nil {
				// disable by default
				rqInternal.Flags.ResetActivation = false

				// ...but enable by default if AlarmClearing is enabled
				if rqInternal.Flags.AlarmClearing {
					rqInternal.Flags.ResetActivation = true
				}
			} else if val {
				// explicit enable
				rqInternal.Flags.ResetActivation = true
			} else {
				rqInternal.Flags.ResetActivation = false
			}

			if val, err := strconv.ParseBool(rqProtocol.Flags.CacheInvalidation); err != nil {
				// enable by default
				rqInternal.Flags.CacheInvalidation = true
			} else if !val {
				// explicit disable
				rqInternal.Flags.CacheInvalidation = false
			} else {
				rqInternal.Flags.CacheInvalidation = true
			}

			if val, err := strconv.ParseBool(rqProtocol.Flags.SendDeploymentFeedback); err != nil {
				// disable by default
				rqInternal.Flags.SendDeploymentFeedback = false
			} else if val {
				// explicit enable
				rqInternal.Flags.SendDeploymentFeedback = true
			} else {
				rqInternal.Flags.SendDeploymentFeedback = false
			}
		}

	case msg.SectionDeployment:
		switch rqInternal.ConfigurationTask {
		case msg.TaskRollout:
			rqInternal.Flags.AlarmClearing = false
			rqInternal.Flags.CacheInvalidation = true
			rqInternal.Flags.ResetActivation = false
			rqInternal.Flags.SendDeploymentFeedback = true
		case msg.TaskDeprovision:
			rqInternal.Flags.AlarmClearing = false
			rqInternal.Flags.CacheInvalidation = false
			rqInternal.Flags.ResetActivation = false
			rqInternal.Flags.SendDeploymentFeedback = true
		case msg.TaskDelete:
			rqInternal.Flags.AlarmClearing = true
			rqInternal.Flags.CacheInvalidation = true
			rqInternal.Flags.ResetActivation = true
			rqInternal.Flags.SendDeploymentFeedback = true
		}
	}
	if rqInternal.Flags.AlarmClearing && !rqInternal.Flags.CacheInvalidation {
		return fmt.Errorf(`Invalid flag combination: alarm.clearing requires cache.invalidation`)
	}
	return nil
}

// foldSlashes collapses sequences of multiple consecutive / characters
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

// stringToTime attempts to parse timestring s into t
func stringToTime(s string, t *time.Time) (err error) {
	for _, format := range []string{
		time.RFC3339,
		eyeproto.RFC3339Milli,
		time.RFC3339Nano,
	} {
		if *t, err = time.Parse(format, s); err == nil {
			// return the string could be successfully parsed
			return
		}
	}
	return
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
