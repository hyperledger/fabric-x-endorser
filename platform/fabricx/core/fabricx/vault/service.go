/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package vault

import (
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/services/logging"
	fdriver "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/storage/vault"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/metrics"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/vault/queryservice"
	"go.opentelemetry.io/otel/trace"
)

var logger = logging.MustGetLogger("fabricx.vault")

func New(
	configService fdriver.ConfigService,
	vaultStore driver.VaultStore,
	channel string,
	adapterProvider protoblocktx.Provider,
	_ queryservice.Provider,
	metricsProvider metrics.Provider,
	tracerProvider trace.TracerProvider,
) (*Vault, error) {
	adapter, err := adapterProvider.Get(configService.NetworkName(), channel)
	if err != nil {
		return nil, err
	}
	// TODO: this is an example how to integrate the query service into the vault and let all getters communicate with the committer directly
	// queryService, err := queryServiceProvider.Get(configService.NetworkName(), channel)
	// if err != nil {
	//	 return nil, nil, fmt.Errorf("could not get mapping provider for %s: %v", channel, err)
	// }
	//
	// wrapp our store with our remote proxy
	// s := queryservice.NewProxyStore(persistence, queryService)

	cachedVault := vault.NewCachedVault(vaultStore, configService.VaultTXStoreCacheSize())
	return NewVault(cachedVault, adapter, metricsProvider, tracerProvider), nil
}
