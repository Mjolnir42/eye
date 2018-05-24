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
func somaStatusUpdate(r *msg.Result) {
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

	baseTimeout := 250
	maxFeedbackAttempts := 5
	client := resty.New()
	success := false

feedbackloop:
	for i := 0; i < maxFeedbackAttempts; i++ {
		<-time.After(time.Duration(i*baseTimeout) * time.Millisecond)

		concurrenyLimit.Start()
		client = client.SetTimeout(time.Duration((i+1)*baseTimeout) * time.Millisecond)
		if res, err := client.R().Patch(url); err == nil {
			concurrenyLimit.Done()

			switch res.StatusCode() {
			case http.StatusOK:
				success = true
				break feedbackloop
			case http.StatusBadRequest:
				// TODO log error, abort: no point resending bad requests
				break feedbackloop
			}
		} else {
			concurrenyLimit.Done()
		}
	}

	log.Println(`RequestID`, r.ID.String(), `DeploymentID`, r.Configuration[0].ID, `Feedback Success`, success)
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
