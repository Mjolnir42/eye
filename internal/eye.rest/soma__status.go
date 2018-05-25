/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty"
	msg "github.com/mjolnir42/eye/internal/eye.msg"
	uuid "github.com/satori/go.uuid"
)

// somaStatusUpdate encapsulates the handling of deployment feedback
// notifications to SOMA
func (x *Rest) somaStatusUpdate(r *msg.Result) {
	if !r.Flags.SendDeploymentFeedback {
		return
	}

	var feedback string

	switch {
	case r.Error != nil:
		feedback = `failed`
	case r.Code >= 400:
		feedback = `failed`
	default:
		feedback = `success`
	}

	url := strings.Replace(r.FeedbackURL, `{STATUS}`, feedback, -1)

	client := resty.New().
		// set generic client options
		SetDisableWarn(true).
		SetHeader(`Content-Type`, `application/json`).
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
		})

	res, err := client.R().Patch(url)
	if err != nil {
		log.Println(`RequestID`, r.ID.String(), `DeploymentID`, r.Configuration[0].ID, `Error`, err.Error())
		return
	}

	switch res.StatusCode() {
	case http.StatusOK:
	default:
		log.Println(`RequestID`, r.ID.String(), `DeploymentID`, r.Configuration[0].ID, res.StatusCode(), res.Status())
	}
}

// somaSetFeedbackURL checks if the SendDeploymentFeedback flag is set
// on r and updates r.FeedbackURL if it is.
func (x *Rest) somaSetFeedbackURL(r *msg.Request) {
	if !r.Flags.SendDeploymentFeedback {
		r.FeedbackURL = ``
		return
	}

	path := x.conf.Eye.SomaPrefix
	feedbackID := r.Configuration.ID

	// potentially better data is available from a SOMA deployment
	// notification
	if !uuid.Equal(uuid.Nil, r.Notification.ID) {
		path = r.Notification.PathPrefix
		feedbackID = r.Notification.ID.String()
	}

	soma, _ := url.Parse(x.conf.Eye.SomaURL)
	soma.Path = fmt.Sprintf("/%s/%s/{STATUS}",
		path,
		feedbackID,
	)
	foldSlashes(soma)
	r.FeedbackURL = soma.String()

}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
