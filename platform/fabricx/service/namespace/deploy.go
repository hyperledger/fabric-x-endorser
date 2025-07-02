/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package namespace

import (
	"encoding/json"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/namespace"
)

type Deploy struct {
	Network   string
	Channel   string
	Namespace string
}

type deployView struct {
	*Deploy
}

func (f *deployView) Call(ctx view.Context) (interface{}, error) {
	deployerService, err := namespace.GetDeployerService(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "deployer service not found")
	}
	return nil, deployerService.DeployNamespace(f.Network, f.Channel, f.Namespace)
}

type DeployViewFactory struct{}

func (p *DeployViewFactory) NewView(in []byte) (view.View, error) {
	f := &deployView{Deploy: &Deploy{}}
	if err := json.Unmarshal(in, f.Deploy); err != nil {
		return nil, err
	}
	return f, nil
}
