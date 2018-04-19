/*-
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/eye/internal/eye.rest"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-resty/resty"
	"github.com/julienschmidt/httprouter"
	msg "github.com/mjolnir42/eye/internal/eye.msg"
	somaproto "github.com/mjolnir42/soma/lib/proto"
	uuid "github.com/satori/go.uuid"
)

// DeploymentNotification implements the API call that receives
// push notifications from SOMA
func (x *Rest) DeploymentNotification(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	// decode client payload
	clientReq := somaproto.NewPushNotification()
	if err := decodeJSONBody(r, &clientReq); err != nil {
		dispatchBadRequest(&w, err.Error())
		return
	}

	// validate client payload
	govalidator.SetFieldsRequiredByDefault(true)
	govalidator.TagMap["abspath"] = govalidator.Validator(func(str string) bool {
		return filepath.IsAbs(str)
	})
	if ok, err := govalidator.ValidateStruct(clientReq); !ok {
		dispatchBadRequest(&w, err.Error())
		return
	}

	// craft internal request message
	request := msg.New(r, params)
	request.Section = msg.SectionDeployment
	request.Action = msg.ActionNotification
	request.Notification = struct {
		ID         uuid.UUID
		PathPrefix string
	}{
		ID:         uuid.FromStringOrNil(clientReq.UUID),
		PathPrefix: clientReq.Path,
	}

	// request authorization for request
	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		soma, _ := url.Parse(x.conf.Eye.SomaURL)
		soma.Path = fmt.Sprintf("/deployments/id/%s/%s", request.Notification.ID.String(), `%s`)
		go sendSomaFeedback(soma.String(), `failed`)
		return
	}

	// fetch DeploymentDetails for ID
	x.fetchPushDeployment(&w, &request)
}

// DeploymentProcess accepts SOMA deployment results
func (x *Rest) DeploymentProcess(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := msg.New(r, params)
	request.Section = msg.SectionDeployment
	request.Action = msg.ActionProcess

	var err error
	cReq := somaproto.NewDeploymentResult()
	if err = decodeJSONBody(r, &cReq); err != nil {
		dispatchInternalError(&w, err)
		return
	}
	if len(*cReq.Deployments) != 1 {
		dispatchUnprocessableEntity(&w, fmt.Errorf("Deployment count %d != 1", len(*cReq.Deployments)))
		return
	}
	request.ConfigurationTask = (*cReq.Deployments)[0].Task
	if request.LookupHash, request.Configuration, err = processDeploymentDetails(&(*cReq.Deployments)[0]); err != nil {
		dispatchInternalError(&w, err)
		return
	}

	// called via v1 update API PUT:/api/v1/item/:ID
	if r.Method == `PUT` {
		if request.Configuration.ID != params.ByName(`ID`) {
			dispatchBadRequest(&w, fmt.Sprintf(
				"Mismatched IDs in update: [%s] vs [%s]",
				request.Configuration.ID,
				params.ByName(`ID`),
			))
			return
		}
	}

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	handler := x.handlerMap.Get(`deployment_w`)
	handler.Intake() <- request
	result := <-request.Reply
	sendMsgResult(&w, &result)
}

// fetchPushDeployment fetches DeploymentDetails for which a push
// notification was received by x.DeploymentNotification
func (x *Rest) fetchPushDeployment(w *http.ResponseWriter, q *msg.Request) {
	var (
		client *resty.Client
		resp   *resty.Response
		res    somaproto.Result
		err    error
	)

	// build URL to download deploymentDetails
	soma, _ := url.Parse(x.conf.Eye.SomaURL)
	soma.Path = strings.Replace(
		fmt.Sprintf("%s/%s",
			q.Notification.PathPrefix,
			q.Notification.ID.String(),
		),
		`//`, `/`,
		-1,
	)
	detailsDownload := soma.String()

	// build URL to send deployment feedback
	soma.Path = fmt.Sprintf("/deployments/id/%s/%s", q.Notification.ID.String(), `%s`)
	q.FeedbackURL = soma.String()

	// fetch DeploymentDetails
	client = resty.New().SetTimeout(750 * time.Millisecond)
	if resp, err = client.R().Get(detailsDownload); err != nil {
		dispatchGatewayTimeout(w, err)
		go sendSomaFeedback(q.FeedbackURL, `failed`)
		return
	}

	// HTTP protocol statuscode > 299
	if resp.StatusCode() > 299 {
		dispatchBadGateway(w, fmt.Errorf("Received: %d/%s", resp.StatusCode(), resp.Status()))
		go sendSomaFeedback(q.FeedbackURL, `failed`)
		return
	}
	if err = json.Unmarshal(resp.Body(), &res); err != nil {
		dispatchInternalError(w, err)
		go sendSomaFeedback(q.FeedbackURL, `failed`)
		return
	}

	// SOMA application statuscode != 200
	if res.StatusCode != 200 {
		dispatchGone(w, fmt.Errorf("SOMA: %d/%s", res.StatusCode, res.StatusText))
		go sendSomaFeedback(q.FeedbackURL, `failed`)
		return
	}

	if len(*res.Deployments) != 1 {
		dispatchUnprocessableEntity(w, fmt.Errorf("Deployment count %d != 1", len(*res.Deployments)))
		go sendSomaFeedback(q.FeedbackURL, `failed`)
		return
	}

	q.ConfigurationTask = (*res.Deployments)[0].Task
	if q.LookupHash, q.Configuration, err = processDeploymentDetails(&(*res.Deployments)[0]); err != nil {
		dispatchInternalError(w, err)
		go sendSomaFeedback(q.FeedbackURL, `failed`)
		return
	}

	if !x.isAuthorized(q) {
		dispatchForbidden(w, nil)
		go sendSomaFeedback(q.FeedbackURL, `failed`)
		return
	}

	handler := x.handlerMap.Get(`deployment_w`)
	handler.Intake() <- *q
	result := <-q.Reply
	sendMsgResult(w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
