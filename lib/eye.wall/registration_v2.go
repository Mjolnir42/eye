/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/solnx/eye/lib/eye.wall"

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-resty/resty"
	proto "github.com/solnx/eye/lib/eye.proto"
	"github.com/solnx/eye/lib/eye.proto/v2"
)

// v2Register implements the cache registration for API version 2
func (l *Lookup) v2Register() error {
	// already registered - unregister first
	if l.registration != `` {
		if err := l.v2Unregister(); err != nil {
			return err
		}
	}

	rq := v2.NewRegistrationRequest()
	rq.Registration = &v2.Registration{
		Application: l.name,
		Address:     strings.Split(l.Config.Redis.Connect, `:`)[0],
		Database:    int64(l.Config.Redis.DB),
	}
	rq.Registration.Port, _ = strconv.ParseInt(
		strings.Split(l.Config.Redis.Connect, `:`)[1],
		10, 64)

	// register cache in Eye
	var err error
	var resp *resty.Response
	var r *v2.Result

	if resp, err = l.client.R().
		SetBody(rq).
		Post(
			l.eyeRegAddURL,
		); err != nil {
		fmt.Println("Error from eye:", err.Error())
		return fmt.Errorf("eyewall.v2Register: %s", err.Error())
	}

	if r, err = v2Result(resp.Body()); err != nil {
		fmt.Println("Error from v2Result:", err.Error())
		return fmt.Errorf("eyewall.v2Register: %s", err.Error())
	}

	// record our cache registrationID
	l.registration = (*r.Registrations)[0].ID

	return nil
}

// v2Unregister implements the cache unregistration for API version 2
func (l *Lookup) v2Unregister() error {
	// not registered
	if l.registration == `` {
		return nil
	}

	var err error
	var resp *resty.Response

	if resp, err = l.client.R().
		SetPathParams(map[string]string{
			`registrationID`: l.registration,
		}).Delete(
		l.eyeRegDelURL,
	); err != nil {
		return fmt.Errorf("eyewall.v2Unregister: %s", err.Error())
	}

	if _, err = v2Result(resp.Body()); err != nil {
		return fmt.Errorf("eyewall.v2Unregister: %s", err.Error())
	}
	l.registration = ``

	return nil
}

// v2LookupRegistrations returns the cache registrations of app via API
// version 2
func (l *Lookup) v2LookupRegistrations(app string) (*proto.Result, error) {
	var err error
	var resp *resty.Response
	var r *v2.Result

	if resp, err = l.client.R().
		SetPathParams(map[string]string{
			`app`: app,
		}).Get(
		l.eyeRegGetURL,
	); err != nil {
		return nil, fmt.Errorf("eyewall.v2LookupRegistrations: %s", err.Error())
	}

	if r, err = v2Result(resp.Body()); err != nil {
		return nil, fmt.Errorf("eyewall.v2LookupRegistrations: %s", err.Error())
	}

	return &proto.Result{
		APIVersion: proto.ProtocolTwo,
		V2Result:   r,
	}, nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
