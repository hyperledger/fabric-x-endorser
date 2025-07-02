/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"fmt"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/utils"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core"
)

type Provider[V any] interface {
	Get(network, channel string) (V, error)
}

type ServiceProvider[V any] struct {
	configService  driver.ConfigService
	defaultNetwork string
	services       map[CommitterVersion]V
}

func NewProvider[V any](configService driver.ConfigService) (*ServiceProvider[V], error) {
	config, err := core.NewConfig(configService)
	if err != nil {
		return nil, err
	}
	return &ServiceProvider[V]{
		configService:  configService,
		defaultNetwork: config.DefaultName(),
		services:       map[CommitterVersion]V{},
	}, nil
}

func (p *ServiceProvider[V]) Register(version CommitterVersion, service V) *ServiceProvider[V] {
	p.services[version] = service
	return p
}

func (p *ServiceProvider[V]) Get(network, _ string) (V, error) {
	if len(network) == 0 {
		network = p.defaultNetwork
	}
	v := p.configService.GetString(fmt.Sprintf("fabric.%s.queryService.version", network))
	if len(v) == 0 {
		return utils.Zero[V](), errors.Errorf("no version defined for network [%s]", network)
	}
	m, ok := p.services[CommitterVersion(v)]
	if !ok {
		return utils.Zero[V](), errors.Errorf("no service defined for version [%s]", v)
	}
	return m, nil
}
