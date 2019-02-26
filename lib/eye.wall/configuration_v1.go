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

	"github.com/go-resty/resty"
	proto "github.com/solnx/eye/lib/eye.proto"
	"github.com/solnx/eye/lib/eye.proto/v1"
)

// v1ConfigurationShow ...
func (l *Lookup) v1ConfigurationShow(profileID string) (*proto.Result, error) {
	var err error
	var resp *resty.Response
	var r *v1.ConfigurationData

	if resp, err = l.client.R().
		SetPathParams(map[string]string{
			`profileID`: profileID,
		}).Get(
		l.eyeCfgGetURL,
	); err != nil {
		return nil, fmt.Errorf("eyewall.v1ConfigurationShow: %s", err.Error())
	}

	if r, err = v1ConfigurationData(resp.Body()); err != nil {
		return nil, fmt.Errorf("eyewall.v1ConfigurationShow: %s", err.Error())
	}

	return &proto.Result{
		APIVersion: proto.ProtocolOne,
		V1Data:     r,
	}, nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
