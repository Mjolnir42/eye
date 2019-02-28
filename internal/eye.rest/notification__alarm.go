/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/solnx/eye/internal/eye.rest"

import (
	"bytes"
	"net/http"
	"time"

	"github.com/go-resty/resty"
	msg "github.com/solnx/eye/internal/eye.msg"
)

func (x *Rest) alarmSend(r *msg.Result) {
	if !r.Flags.AlarmClearing {
		return
	}

configurationloop:
	for i := range r.Configuration {
		snap := r.Configuration[i].At(r.Time)
		if !snap.Valid {
			continue configurationloop
		}

		body := &bytes.Buffer{}
		x.tmpl.Execute(body, snap)
		go func(b []byte, uri string, mr *msg.Result, idx int) {
			res, err := resty.New().
				// set generic client options
				SetDisableWarn(true).
				SetHeader(`Content-Type`, x.conf.Eye.AlarmContentType).
				SetContentLength(true).
				// follow redirects
				SetRedirectPolicy(resty.FlexibleRedirectPolicy(5)).
				// configure request retry
				SetRetryCount(x.conf.Eye.RetryCount).
				SetRetryWaitTime(time.Duration(x.conf.Eye.RetryMinWaitTime) * time.Millisecond).
				SetRetryMaxWaitTime(time.Duration(x.conf.Eye.RetryMaxWaitTime) * time.Millisecond).
				// reset timeout deadline before every request
				OnBeforeRequest(func(cl *resty.Client, rq *resty.Request) error {
					cl.SetTimeout(time.Duration(x.conf.Eye.RequestTimeout) * time.Millisecond)
					return nil
				}).
				// enter concurrency limit before performing request
				OnBeforeRequest(func(cl *resty.Client, rq *resty.Request) error {
					x.limit.Start()
					return nil
				}).
				// leave concurrency limit after receiving a response
				OnAfterResponse(func(cl *resty.Client, rp *resty.Response) error {
					x.limit.Done()
					return nil
				}).
				// clear timeout deadline after each request (http.Client
				// timeout also cancels reading the response body)
				OnAfterResponse(func(cl *resty.Client, rp *resty.Response) error {
					cl.SetTimeout(0)
					return nil
				}).
				R().
				SetBody(b).
				Post(uri)

			if err != nil {

				x.appLog.Errorf(`Error clearing alarm for RequestID: %s DeploymentID: %s Error: %s`, mr.ID.String(), mr.Configuration[idx].ID, err.Error())
				return
			}
			switch res.StatusCode() {
			case http.StatusOK:
				x.appLog.Debugln(`Alarm clearing for RequestID: %s DeploymentID: %s Status: %s`, mr.ID.String(), mr.Configuration[idx].ID, res.Status())
			default:
				x.appLog.Errorln(`Invalid status on alarm clearing for RequestID: %s DeploymentID: %s Status: %s`, mr.ID.String(), mr.Configuration[idx].ID, res.Status())
			}
		}(body.Bytes(), x.conf.Eye.AlarmEndpoint, r, i)
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
