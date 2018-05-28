/*-
 * Copyright © 2016,2017, Jörg Pernfuß <code.jpe@gmail.com>
 * Copyright © 2016, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

// Package wall provides a lookup library for threshold
// configurations managed by eye
package wall // import "github.com/mjolnir42/eye/lib/eye.wall"

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/go-resty/resty"
	"github.com/mjolnir42/erebos"
	proto "github.com/mjolnir42/eye/lib/eye.proto"
	"github.com/mjolnir42/limit"
)

var (
	// ErrNotFound is returned when the cache contains no matching data
	ErrNotFound = errors.New("eyewall.Lookup: not found")
	// ErrUnconfigured is returned when the cache contains a negative
	// caching entry or Eye returns the absence of a profile to look up
	ErrUnconfigured = errors.New("eyewall.Lookup: unconfigured")
	// ErrUnavailable is returned when the cache does not contain the
	// requested record and Eye can not be queried
	ErrUnavailable = errors.New(`eyewall.Lookup: profile server unavailable`)
	// beats is the map of heartbeats shared between all instances of
	// Lookup. This way it can be ensured that all instances only move
	// the timestamps forward in time.
	beats heartbeatMap
)

func init() {
	beats.hb = make(map[int]time.Time)
}

// Lookup provides a query library to retrieve data from Eye
type Lookup struct {
	Config       *erebos.Config
	limit        *limit.Limit
	log          *logrus.Logger
	redis        *redis.Client
	cacheTimeout time.Duration
	apiVersion   int
}

// NewLookup returns a new *Lookup
func NewLookup(conf *erebos.Config) *Lookup {
	return &Lookup{
		Config: conf,
		limit:  limit.New(conf.Eyewall.ConcurrencyLimit),
		log:    nil,
	}
}

// Start sets up Lookup and connects to Redis
func (l *Lookup) Start() error {
	l.Taste()
	l.cacheTimeout = time.Duration(
		l.Config.Redis.CacheTimeout,
	) * time.Second
	l.redis = redis.NewClient(&redis.Options{
		Addr:     l.Config.Redis.Connect,
		Password: l.Config.Redis.Password,
		DB:       l.Config.Redis.DB,
	})
	if _, err := l.redis.Ping().Result(); err != nil {
		return err
	}
	return nil
}

// Close shuts down the Redis connection
func (l *Lookup) Close() {
	l.redis.Close()
}

// Taste connects to Eye and checks supported API versions
func (l *Lookup) Taste() {
	// use high timeout + retry variant for initial tasting
	l.taste(false)
}

// taste connects to Eye and checks supported API versions. If quick is
// set, no retries and shorter timeouts are used.
func (l *Lookup) taste(quick bool) {
	var retryCount, timeoutMS int
	switch quick {
	case true:
		retryCount = 0
		timeoutMS = 150
	case false:
		retryCount = 2
		timeoutMS = 250
	}

versionloop:
	// protocol version array is preference sorted, first hit wins
	for _, apiVersion := range []int{proto.ProtocolTwo, proto.ProtocolOne} {
		eyeURL, err := url.Parse(fmt.Sprintf("http://%s:%s/api?version=%d",
			l.Config.Eyewall.Host,
			l.Config.Eyewall.Port,
			apiVersion,
		))
		if err != nil {
			l.log.Fatalf("eyewall/cache: malformed eye URL: %s", err.Error())
		}

		foldSlashes(eyeURL)

		resp, err := resty.New().
			// set generic client options
			SetHeader(`Content-Type`, `application/json`).
			SetContentLength(true).
			// follow redirects
			SetRedirectPolicy(resty.FlexibleRedirectPolicy(5)).
			// configure request retry
			SetRetryCount(retryCount).
			SetRetryWaitTime(500 * time.Millisecond).
			SetRetryMaxWaitTime(3000 * time.Millisecond).
			// reset timeout deadline before every request
			OnBeforeRequest(func(cl *resty.Client, rq *resty.Request) error {
				cl.SetTimeout(time.Duration(timeoutMS) * time.Millisecond)
				return nil
			}).
			// enter concurrency limit before performing request
			OnBeforeRequest(func(cl *resty.Client, rq *resty.Request) error {
				l.limit.Start()
				return nil
			}).
			// leave concurrency limit after receiving a response
			OnAfterResponse(func(cl *resty.Client, rp *resty.Response) error {
				l.limit.Done()
				return nil
			}).
			// clear timeout deadline after each request (http.Client
			// timeout also cancels reading the response body)
			OnAfterResponse(func(cl *resty.Client, rp *resty.Response) error {
				cl.SetTimeout(0)
				return nil
			}).
			R().Head(eyeURL.String())

		// connection error to eye is NOT fatal, run against cache instead
		if err != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
			break versionloop // stop trying to find a different api version
		}

		switch resp.StatusCode() {
		case http.StatusBadRequest: // 400
			// unfixable at runtime and eyewall can not work without
			// figuring out the API version
			panic(`Eyewall received BadRequest response from Eye in response to API version testing. Can not continue.`)
		case http.StatusNotImplemented: // 501
			// queried protocol version is not supported
			continue versionloop
		case http.StatusNotFound: // 404
			// eye is so old, it does not have the /api endpoint.
			// This means it is only able to handle ProtocolOne
			l.apiVersion = proto.ProtocolOne
			break versionloop
		case http.StatusNoContent: // 204
			// queried version is supported
			l.apiVersion = apiVersion
			break versionloop
		}
	}
}

// SetLogger hands Lookup the logger to use
func (l *Lookup) SetLogger(logger *logrus.Logger) {
	l.log = logger
}

// GetConfigurationID returns matching monitoring profile ConfigurationIDs
// if any exist.
func (l *Lookup) GetConfigurationID(lookID string) ([]string, error) {
	IDList := []string{}

	// try to serve the request from the local redis cache
	thresh, err := l.processRequest(lookID)
	switch err {
	case nil:
		// success
	case ErrUnconfigured:
		// lookID has negative cache entry
		return IDList, ErrUnconfigured
	default:
		// genuine error condition
		return IDList, err
	}

	for k := range thresh {
		IDList = append(IDList, thresh[k].ID)
	}
	return IDList, nil
}

// LookupThreshold queries the full monitoring profile data
// for lookID
func (l *Lookup) LookupThreshold(lookID string) (map[string]Threshold, error) {
	return l.processRequest(lookID)
}

// WaitEye returns a channel that it closes once Eye returns a
// valid result without errors
func (l *Lookup) WaitEye() chan struct{} {
	ret := make(chan struct{})
	go func(ret chan struct{}) {
		retryDelay := 50 * time.Millisecond
		client := &http.Client{}
		for {
			<-time.After(retryDelay)
			retryDelay = 5 * time.Second

			req, err := http.NewRequest(`GET`, fmt.Sprintf(
				"http://%s:%s/api/v1/item/",
				l.Config.Eyewall.Host,
				l.Config.Eyewall.Port,
			), nil)
			if err != nil {
				continue
			}
			var resp *http.Response
			if resp, err = client.Do(req); err != nil {
				ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				continue
			}
			// allow connection reuse
			ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				break
			}
		}
		close(ret)
	}(ret)
	return ret
}

// processRequest handles the multi-stage lookup of querying the
// cache, the profile server and keeps the cache updated
func (l *Lookup) processRequest(lookID string) (map[string]Threshold, error) {
	// fetch from local cache
	thr, err := l.lookupRedis(lookID)
	if err == nil {
		return thr, nil
	} else if err == ErrUnconfigured {
		return nil, ErrUnconfigured
	} else if err != ErrNotFound {
		// genuine error condition, defer to profile server
		if l.log != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
		}
	}

	// local cache did not hit or was not available
	// fetch from eye

	// apiVersion is not initialized, run a quick tasting
	if l.apiVersion == proto.ProtocolInvalid {
		l.taste(true)
	}

	switch l.apiVersion {
	case proto.ProtocolInvalid:
		// apiVersion is still uninitialized, this is now a hard error
		// since the cache does not have the required data and eye can
		// not be queried
		return nil, ErrUnavailable
	case proto.ProtocolOne:
		cnf, err := l.v1LookupEye(lookID)
		if err == ErrUnconfigured {
			return nil, ErrUnconfigured
		} else if err != nil {
			return nil, err
		}

		// process result from eye and store in redis
		thr, err = l.v1Process(lookID, cnf)
		if err == ErrUnconfigured {
			return nil, ErrUnconfigured
		} else if err != nil {
			return nil, err
		}
	case proto.ProtocolTwo:
	default:
	}

	return thr, nil
}

// lookupRedis queries the Redis profile cache
func (l *Lookup) lookupRedis(lookID string) (map[string]Threshold, error) {
	res := make(map[string]Threshold)
	data, err := l.redis.HGetAll(lookID).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrNotFound
	}
dataloop:
	for key := range data {
		if key == `unconfigured` {
			if len(data) == 1 {
				return nil, ErrUnconfigured
			}
			continue dataloop
		}
		val, err := l.redis.Get(key).Result()
		if err != nil {
			return nil, err
		}

		t := Threshold{}
		err = json.Unmarshal([]byte(val), &t)
		if err != nil {
			return nil, err
		}
		res[t.ID] = t
	}
	return res, nil
}

// setUnconfigured writes a negative cache entry into the local cache
func (l *Lookup) setUnconfigured(lookID string) {
	if _, err := l.redis.HSet(
		lookID,
		`unconfigured`,
		time.Now().UTC().Format(time.RFC3339),
	).Result(); err != nil {
		if l.log != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
		}
		return
	}

	if _, err := l.redis.Expire(
		lookID,
		l.cacheTimeout,
	).Result(); err != nil {
		if l.log != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
		}
	}
}

// storeThreshold writes t into the local cache
func (l *Lookup) storeThreshold(lookID string, t *Threshold) {
	buf, err := json.Marshal(t)
	if err != nil {
		if l.log != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
		}
		return
	}

	if _, err := l.redis.Set(
		t.ID,
		string(buf),
		l.cacheTimeout,
	).Result(); err != nil {
		if l.log != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
		}
		return
	}

	if _, err := l.redis.HSet(
		lookID,
		t.ID,
		time.Now().UTC().Format(time.RFC3339),
	).Result(); err != nil {
		if l.log != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
		}
		return
	}

	if _, err := l.redis.Expire(
		lookID,
		l.cacheTimeout,
	).Result(); err != nil {
		if l.log != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
		}
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
