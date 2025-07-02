/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricx

import (
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic"
	driver2 "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/driver/config"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/driver/identity"
	fdriver "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/kvs"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/metrics"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/transaction"
)

type Provider struct {
	p               *driver2.Provider
	adapterProvider protoblocktx.Provider
}

func NewProvider(
	configProvider config.Provider,
	metricsProvider metrics.Provider,
	endpointService identity.EndpointService,
	channelProvider generic.ChannelProvider,
	idProvider identity.ViewIdentityProvider,
	identityLoaders []identity.NamedIdentityLoader,
	kvss *kvs.KVS,
	signerKVS driver.SignerInfoStore,
	auditInfoKVS driver.AuditInfoStore,
	adapterProvider protoblocktx.Provider,
) *Provider {
	return &Provider{
		p: driver2.NewProvider(
			configProvider,
			metricsProvider,
			endpointService,
			channelProvider,
			idProvider,
			identityLoaders,
			signerKVS,
			auditInfoKVS,
			kvss,
		),
		adapterProvider: adapterProvider,
	}
}

func (d *Provider) New(network string, b bool) (fdriver.FabricNetworkService, error) {
	net, err := d.p.New(network, b)
	if err != nil {
		return nil, err
	}

	adapter, err := d.adapterProvider.Get(network, "")
	if err != nil {
		return nil, err
	}
	txManager := transaction.NewManager(adapter)
	txManager.AddTransactionFactory(
		fdriver.EndorserTransaction,
		transaction.NewTransactionFactory(net, adapter),
	)

	net.(*generic.Network).SetTransactionManager(txManager)

	return net, nil
}
