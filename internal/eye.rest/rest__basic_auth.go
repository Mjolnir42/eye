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
	"encoding/base64"

	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	uuid "github.com/satori/go.uuid"
	"github.com/solnx/eye/internal/eye"
	msg "github.com/solnx/eye/internal/eye.msg"
)

func (x *Rest) EnrichRequest(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request,
		ps httprouter.Params) {
		// generate and record the requestID
		requestID := uuid.Must(uuid.NewV4())
		ps = append(ps, httprouter.Param{
			Key:   `RequestID`,
			Value: requestID.String(),
		})
		requestTS := time.Now().UTC()
		ps = append(ps, httprouter.Param{
			Key:   `RequestTS`,
			Value: requestTS.Format(time.RFC3339Nano),
		})
		h(w, r, ps)
		return
	}
}

// BasicAuth handles HTTP BasicAuth on requests
func (x *Rest) BasicAuth(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request,
		ps httprouter.Params) {
		const basicAuthPrefix string = "Basic "
		var supervisor eye.Handler

		// generate and record the requestID
		requestID := uuid.Must(uuid.NewV4())
		ps = append(ps, httprouter.Param{
			Key:   `RequestID`,
			Value: requestID.String(),
		})
		requestTS := time.Now().UTC()
		ps = append(ps, httprouter.Param{
			Key:   `RequestTS`,
			Value: requestTS.Format(time.RFC3339Nano),
		})

		// if the supervisor is not available, no requests are accepted
		if supervisor = x.handlerMap.Get(`supervisor`); supervisor == nil {
			http.Error(w, `Authentication supervisor not available`,
				http.StatusServiceUnavailable)
			return
		}

		// v1 API was unauthenticated
		request := msg.New(r, ps)
		if request.Version == msg.ProtocolOne {
			// record fake authentication information
			ps = append(ps, httprouter.Param{
				Key:   `AuthenticatedUser`,
				Value: `nobody`,
			})
			ps = append(ps, httprouter.Param{
				Key:   `AuthenticatedToken`,
				Value: `v1apirequest`,
			})
			// Delegate request to given handle
			h(w, r, ps)
			return
		}

		// Get credentials
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, basicAuthPrefix) {
			// Check credentials
			payload, err := base64.StdEncoding.DecodeString(
				auth[len(basicAuthPrefix):],
			)
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 {
					request.Section = msg.SectionSupervisor
					request.Action = msg.ActionAuthenticate
					request.Super = msg.Supervisor{
						Task: msg.TaskBasicAuth,
						BasicAuth: struct {
							User  []byte
							Token []byte
						}{
							User:  pair[0],
							Token: pair[1],
						},
					}
					supervisor.Intake() <- request

					result := <-request.Reply
					if result.Error != nil {
						x.appLog.Errorln(result.Error.Error()) // XXX
					}

					if result.Super.Verdict == msg.VerdictOK {
						// record the authenticated user
						ps = append(ps, httprouter.Param{
							Key:   `AuthenticatedUser`,
							Value: string(pair[0]),
						})
						// record the used token
						ps = append(ps, httprouter.Param{
							Key:   `AuthenticatedToken`,
							Value: string(pair[1]),
						})
						// Delegate request to given handle
						h(w, r, ps)
						return
					}
				}
			}
		}

		w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		http.Error(w, http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
