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
package wall // import "github.com/solnx/eye/lib/eye.wall"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/go-resty/resty"
	"github.com/mjolnir42/erebos"
	"github.com/mjolnir42/limit"
	"github.com/solnx/eye/internal/eye.msg"
	proto "github.com/solnx/eye/lib/eye.proto"
	"github.com/solnx/eye/lib/eye.proto/v2"
)

var (
	// beats is the map of heartbeats shared between all instances of
	// Lookup. This way it can be ensured that all instances only move
	// the timestamps forward in time.
	beats heartbeatMap
)

func init() {
	beats.hb = make(map[int]time.Time)

	// set timestamp formatting options
	v2.TimeFormatString = proto.RFC3339Milli
	v2.PosTimeInf = msg.PosTimeInf
	v2.NegTimeInf = msg.NegTimeInf
}

// Lookup provides a query library to retrieve data from Eye
type Lookup struct {
	Config       *erebos.Config
	limit        *limit.Limit
	log          *logrus.Logger
	redis        *redis.Client
	pipe         redis.Pipeliner
	cacheTimeout time.Duration
	apiVersion   int
	eyeLookupURL string
	eyeActiveURL string
	eyeRegAddURL string
	eyeRegDelURL string
	eyeRegGetURL string
	eyeCfgGetURL string
	eyeActPndURL string
	client       *resty.Client
	name         string
	registration string
}

// NewLookup returns a new *Lookup
func NewLookup(conf *erebos.Config, appName string) *Lookup {
	l := &Lookup{
		Config: conf,
		limit:  limit.New(conf.Eyewall.ConcurrencyLimit),
		log:    nil,
		name:   appName,
	}
	// use configured application name if it was set
	if l.Config.Eyewall.ApplicationName != `` {
		l.name = l.Config.Eyewall.ApplicationName
	}
	// set defaults for the connections pools if nothing else is specified
	if l.Config.Eyewall.ConnectionPool == 0 {
		l.Config.Eyewall.ConnectionPool = 500
	}
	if l.Config.Redis.PoolSize == 0 {
		l.Config.Redis.PoolSize = 20
	}
	if l.Config.Redis.MinIdleConns == 0 {
		l.Config.Redis.PoolSize = 10
	}
	transport := &http.Transport{
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConnsPerHost: l.Config.Eyewall.ConnectionPool,
		MaxIdleConns:        l.Config.Eyewall.ConnectionPool,
	}
	l.client = resty.New().
		SetHeader(`Content-Type`, `application/json`).
		SetContentLength(true).
		SetDisableWarn(true).
		SetTransport(transport).
		SetRedirectPolicy(resty.NoRedirectPolicy()).
		SetRetryCount(0).
		OnBeforeRequest(func(cl *resty.Client, rq *resty.Request) error {
			cl.SetTimeout(150 * time.Millisecond)
			return nil
		}).
		OnBeforeRequest(func(cl *resty.Client, rq *resty.Request) error {
			l.limit.Start()
			return nil
		}).
		OnAfterResponse(func(cl *resty.Client, rp *resty.Response) error {
			l.limit.Done()
			return nil
		}).
		OnAfterResponse(func(cl *resty.Client, rp *resty.Response) error {
			cl.SetTimeout(0)
			return nil
		})
	return l
}

// Start sets up Lookup and connects to Redis
func (l *Lookup) Start() error {
	l.Taste()

	if l.Config.Eyewall.NoLocalRedis {
		l.log.Debugf("Started new eyewall instance ConnectionPool=%d Redis=Disabled", l.Config.Eyewall.ConnectionPool)
		return nil
	}
	l.log.Debugf("Started new eyewall instance ConnectionPool=%d RedisPool=%d", l.Config.Eyewall.ConnectionPool, l.Config.Redis.PoolSize)
	l.cacheTimeout = time.Duration(
		l.Config.Redis.CacheTimeout,
	) * time.Second
	l.redis = redis.NewClient(&redis.Options{
		Addr:         l.Config.Redis.Connect,
		Password:     l.Config.Redis.Password,
		DB:           l.Config.Redis.DB,
		PoolSize:     500,
		MinIdleConns: 10,
	})
	l.pipe = l.redis.Pipeline()
	if _, err := l.redis.Ping().Result(); err != nil {
		return err
	}
	if err := l.resetReceived(); err != nil {
		return err
	}
	return l.Register()
}

// Close shuts down the Redis connection
func (l *Lookup) Close() {
	if l.Config.Eyewall.NoLocalRedis {
		return
	}
	l.redis.Close()
	if err := l.Unregister(); err != nil {
		l.log.Errorf("Error on Lookup Unregister: %s", err.Error())
	}
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
			SetDisableWarn(true).
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
	switch l.apiVersion {
	case proto.ProtocolOne:
		l.eyeLookupURL = fmt.Sprintf("http://%s:%s/api/v1/configuration/{lookID}",
			l.Config.Eyewall.Host,
			l.Config.Eyewall.Port,
		)

		l.eyeCfgGetURL = fmt.Sprintf("http://%s:%s/api/v1/item/{profileID}",
			l.Config.Eyewall.Host,
			l.Config.Eyewall.Port,
		)
	case proto.ProtocolTwo:
		l.eyeLookupURL = fmt.Sprintf("http://%s:%s/api/v2/lookup/configuration/{lookID}",
			l.Config.Eyewall.Host,
			l.Config.Eyewall.Port,
		)

		l.eyeActiveURL = fmt.Sprintf("http://%s:%s/api/v2/configuration/{profileID}/active",
			l.Config.Eyewall.Host,
			l.Config.Eyewall.Port,
		)

		l.eyeRegAddURL = fmt.Sprintf("http://%s:%s/api/v2/registration/",
			l.Config.Eyewall.Host,
			l.Config.Eyewall.Port,
		)

		l.eyeRegDelURL = fmt.Sprintf("http://%s:%s/api/v2/registration/{registrationID}",
			l.Config.Eyewall.Host,
			l.Config.Eyewall.Port,
		)

		l.eyeRegGetURL = fmt.Sprintf("http://%s:%s/api/v2/lookup/registration/{app}",
			l.Config.Eyewall.Host,
			l.Config.Eyewall.Port,
		)

		l.eyeCfgGetURL = fmt.Sprintf("http://%s:%s/api/v2/configuration/{profileID}",
			l.Config.Eyewall.Host,
			l.Config.Eyewall.Port,
		)

		l.eyeActPndURL = fmt.Sprintf("http://%s:%s/api/v2/lookup/activation/",
			l.Config.Eyewall.Host,
			l.Config.Eyewall.Port,
		)
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
		res, err := l.v2LookupEye(lookID)
		if err == ErrUnconfigured {
			return nil, ErrUnconfigured
		} else if err != nil {
			return nil, err
		}

		// process result from eye and store in redis
		thr, err = l.v2Process(lookID, res)
		if err == ErrUnconfigured {
			return nil, ErrUnconfigured
		} else if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("eyewall.Lookup: attempted processing for unsupported API version %d", l.apiVersion)
	}

	return thr, nil
}

// lookupRedis queries the Redis profile cache
func (l *Lookup) lookupRedis(lookID string) (map[string]Threshold, error) {
	if l.Config.Eyewall.NoLocalRedis {
		return nil, ErrNoCache
	}

	res := make(map[string]Threshold)

	getall := l.pipe.HGetAll(lookID)
	_, err := l.pipe.Exec()
	if err != nil {
		return nil, err
	}
	data := getall.Val()
	if len(data) == 0 {
		return nil, ErrNotFound
	}
	getMap := make(map[string]*redis.StringCmd)

dataloop:
	for key := range data {
		if key == `unconfigured` {
			if len(data) == 1 {
				return nil, ErrUnconfigured
			}
			continue dataloop
		}
		getMap[key] = l.pipe.Get(key)
	}
	_, err = l.pipe.Exec()
	if err != nil {
		return nil, err
	}
	for key := range getMap {
		t := Threshold{}
		err = json.Unmarshal([]byte(getMap[key].Val()), &t)
		if err != nil {
			return nil, err
		}
		res[t.ID] = t
	}

	return res, nil
}

// setUnconfigured writes a negative cache entry into the local cache
func (l *Lookup) setUnconfigured(lookID string) {
	if l.Config.Eyewall.NoLocalRedis {
		return
	}
	l.pipe.HSet(
		lookID,
		`unconfigured`,
		time.Now().UTC().Format(time.RFC3339),
	)

	l.pipe.Expire(
		lookID,
		l.cacheTimeout,
	)
	//	_, err := pipe.Exec()
	//	if err != nil {
	//		if l.log != nil {
	//			l.log.Errorf("eyewall/cache: %s", err.Error())
	//		}
	//	}
}

// storeThreshold writes t into the local cache
func (l *Lookup) storeThreshold(lookID string, t *Threshold) {
	if l.Config.Eyewall.NoLocalRedis {
		return
	}

	buf, err := json.Marshal(t)
	if err != nil {
		if l.log != nil {
			l.log.Errorf("eyewall/cache: %s", err.Error())
		}
		return
	}

	l.pipe.Set(
		t.ID,
		string(buf),
		l.cacheTimeout,
	)

	l.pipe.HSet(
		lookID,
		t.ID,
		time.Now().UTC().Format(time.RFC3339),
	)

	l.pipe.Expire(
		lookID,
		l.cacheTimeout,
	)
	//	_, err = pipe.Exec()
	//	if err != nil {
	//		if l.log != nil {
	//			l.log.Errorf("eyewall/cache: %s", err.Error())
	//		}
	//	}
}

// APIVersion returns the Eye API version discovered via taste testing
func (l *Lookup) APIVersion() int {
	return l.apiVersion
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
