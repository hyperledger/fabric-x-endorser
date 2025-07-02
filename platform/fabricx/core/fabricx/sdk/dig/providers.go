/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/core/generic/committer"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic"
	fcommitter "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/committer"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/delivery"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/driver/config"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/driver/identity"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/membership"
	vault2 "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/vault"
	fdriver "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/services/db/driver/multiplexed"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/events"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/grpc"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/hash"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/kvs"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/metrics"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx"
	committer2 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/ledger"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/transaction/rwset"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/vault"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/vault/queryservice"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

const FabricxDriverName = "fabricx"

func NewDriver(in struct {
	dig.In
	ConfigProvider  config.Provider
	MetricsProvider metrics.Provider
	EndpointService identity.EndpointService
	IdProvider      identity.ViewIdentityProvider
	KVS             *kvs.KVS
	SignerInfoStore driver.SignerInfoStore
	AuditInfoStore  driver.AuditInfoStore
	AdapterProvider protoblocktx.Provider
	ChannelProvider ChannelProvider
	IdentityLoaders []identity.NamedIdentityLoader `group:"identity-loaders"`
},
) core.NamedDriver {
	d := core.NamedDriver{
		Name: FabricxDriverName,
		Driver: fabricx.NewProvider(
			in.ConfigProvider,
			in.MetricsProvider,
			in.EndpointService,
			in.ChannelProvider,
			in.IdProvider,
			in.IdentityLoaders,
			in.KVS,
			in.SignerInfoStore,
			in.AuditInfoStore,
			in.AdapterProvider,
		),
	}
	return d
}

type ChannelProvider generic.ChannelProvider

func NewChannelProvider(in struct {
	dig.In
	ConfigProvider          config.Provider
	KVS                     *kvs.KVS
	LedgerProvider          ledger.Provider
	Publisher               events.Publisher
	BlockDispatcherProvider *ledger.BlockDispatcherProvider
	Hasher                  hash.Hasher
	TracerProvider          trace.TracerProvider
	MetricsProvider         metrics.Provider
	AdapterProvider         protoblocktx.Provider
	QueryServiceProvider    queryservice.Provider
	IdentityLoaders         []identity.NamedIdentityLoader `group:"identity-loaders"`
	EndpointService         identity.EndpointService
	IdProvider              identity.ViewIdentityProvider
	EnvelopeStore           fdriver.EnvelopeStore
	MetadataStore           fdriver.MetadataStore
	EndorseTxStore          fdriver.EndorseTxStore
	Drivers                 multiplexed.Driver
},
) generic.ChannelProvider {
	channelConfigProvider := generic.NewChannelConfigProvider(in.ConfigProvider)
	flmProvider := committer.NewFinalityListenerManagerProvider[fdriver.ValidationCode](in.TracerProvider)
	return generic.NewChannelProvider(
		in.ConfigProvider,
		in.EnvelopeStore,
		in.MetadataStore,
		in.EndorseTxStore,
		in.Hasher,
		in.Drivers,
		func(channelName string, configService fdriver.ConfigService, vaultStore driver.VaultStore) (*vault2.Vault, error) {
			return vault.New(configService, vaultStore, channelName, in.AdapterProvider, in.QueryServiceProvider, in.MetricsProvider, in.TracerProvider)
		},
		channelConfigProvider,

		func(channelName string, nw fdriver.FabricNetworkService, chaincodeManager fdriver.ChaincodeManager) (fdriver.Ledger, error) {
			return in.LedgerProvider.NewLedger(nw.Name(), channelName)
		},
		func(channel string, nw fdriver.FabricNetworkService, envelopeService fdriver.EnvelopeService, transactionService fdriver.EndorserTransactionService, vault fdriver.RWSetInspector) (fdriver.RWSetLoader, error) {
			return NewRWSetLoader(channel, nw, envelopeService, transactionService, vault), nil
		}, func(nw fdriver.FabricNetworkService, channel string, vault fdriver.Vault, envelopeService fdriver.EnvelopeService, ledger fdriver.Ledger, rwsetLoaderService fdriver.RWSetLoader, channelMembershipService *membership.Service, fabricFinality fcommitter.FabricFinality, quiet bool) (generic.CommitterService, error) {
			channelConfig, err := channelConfigProvider.GetChannelConfig(nw.Name(), channel)
			if err != nil {
				return nil, err
			}
			return NewCommitter(nw, channelConfig, vault, envelopeService, ledger, rwsetLoaderService, in.Publisher, channelMembershipService, fabricFinality, fcommitter.NewSerialDependencyResolver(), quiet, flmProvider.NewManager(), in.TracerProvider, in.MetricsProvider)
		},
		// delivery service constructor
		func(
			nw fdriver.FabricNetworkService,
			channel string,
			peerManager delivery.Services,
			ledger fdriver.Ledger,
			vault delivery.Vault,
			callback fdriver.BlockCallback,
		) (generic.DeliveryService, error) {
			// we inject here the block dispatcher and the callback
			// note that once the committer queryservice/notification service is available, we will remove the
			// local commit-pipeline and delivery service
			dispatcher, err := in.BlockDispatcherProvider.GetBlockDispatcher(nw.Name(), channel)
			if err != nil {
				return nil, err
			}
			channelConfig, err := channelConfigProvider.GetChannelConfig(nw.Name(), channel)
			if err != nil {
				return nil, err
			}
			dispatcher.AddCallback(callback)

			return delivery.NewService(
				channel,
				channelConfig,
				in.Hasher,
				nw.Name(),
				nw.LocalMembership(),
				nw.ConfigService(),
				peerManager,
				ledger,
				vault,
				nw.TransactionManager(),
				dispatcher.OnBlock,
				in.TracerProvider,
				in.MetricsProvider,
				[]common.HeaderType{common.HeaderType_MESSAGE},
			)
		},
		false,
	)
}

func NewRWSetLoader(channel string, nw fdriver.FabricNetworkService, envelopeService fdriver.EnvelopeService, transactionService fdriver.EndorserTransactionService, vault fdriver.RWSetInspector) fdriver.RWSetLoader {
	return rwset.NewLoader(nw.Name(), channel, envelopeService, transactionService, nw.TransactionManager(), vault)
}

func NewCommitter(nw fdriver.FabricNetworkService, channelConfig fdriver.ChannelConfig, vault fdriver.Vault, envelopeService fdriver.EnvelopeService, ledger fdriver.Ledger, rwsetLoaderService fdriver.RWSetLoader, eventsPublisher events.Publisher, channelMembershipService *membership.Service, fabricFinality fcommitter.FabricFinality, dependencyResolver fcommitter.DependencyResolver, quiet bool, listenerManager fdriver.ListenerManager, tracerProvider trace.TracerProvider, metricsProvider metrics.Provider) (*fcommitter.Committer, error) {
	os, ok := nw.OrderingService().(fcommitter.OrderingService)
	if !ok {
		return nil, errors.New("ordering service is not a committer.OrderingService")
	}
	c := fcommitter.New(
		nw.ConfigService(),
		channelConfig,
		vault,
		envelopeService,
		ledger,
		rwsetLoaderService,
		nw.ProcessorManager(),
		eventsPublisher,
		channelMembershipService,
		os,
		fabricFinality,
		nw.TransactionManager(),
		dependencyResolver,
		quiet,
		listenerManager,
		tracerProvider,
		metricsProvider,
	)

	// consider meta namespace transactions to be stored in the vault
	if err := c.ProcessNamespace(protoblocktx.MetaNamespace); err != nil {
		return nil, err
	}

	// register fabricx transaction handler
	committer2.RegisterTransactionHandler(c)
	return c, nil
}

func NewConfigProvider(p config.Provider) config.Provider {
	return &configProvider{Provider: p}
}

type configProvider struct {
	config.Provider
}

func (p *configProvider) GetConfig(network string) (config.ConfigService, error) {
	c, err := p.Provider.GetConfig(network)
	if err != nil {
		return nil, err
	}
	var peers []*grpc.ConnectionConfig
	if err := c.UnmarshalKey("peers", &peers); err != nil {
		return nil, err
	}

	logger.Debugf("Getting config for [%s] network; found %d peers", network, len(peers))
	if logger.IsEnabledFor(zap.DebugLevel) {
		for _, p := range peers {
			logger.Debugf("Peer [%s]", p.Address)
		}
	}

	return &configService{ConfigService: c, peers: peers}, nil
}

type configService struct {
	config.ConfigService
	peers []*grpc.ConnectionConfig
}

func (s *configService) PickPeer(fdriver.PeerFunctionType) *grpc.ConnectionConfig {
	logger.Infof("Picking peer: [%s]", s.peers[len(s.peers)-1].Address)
	return s.peers[len(s.peers)-1]
}
