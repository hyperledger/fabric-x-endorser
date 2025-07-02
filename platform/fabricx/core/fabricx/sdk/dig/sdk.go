/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"context"
	"errors"

	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	common "github.com/hyperledger-labs/fabric-smart-client/platform/common/sdk/dig"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/services/logging"
	digutils "github.com/hyperledger-labs/fabric-smart-client/platform/common/utils/dig"
	fabric "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/sdk/dig"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoqueryservice"
	v1 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v1"
	protoblocktx2 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v1/protoblocktx"
	protoqueryservice2 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v1/protoqueryservice"
	v2 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v2"
	protoblocktx3 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v2/protoblocktx"
	protoqueryservice3 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v2/protoqueryservice"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/finality"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/ledger"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/namespace"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/vault/queryservice"
	"go.uber.org/dig"
)

var logger = logging.MustGetLogger("fabricx.sdk")

// SDK extends the fabric SDK with fabricX.
type SDK struct {
	common.SDK
}

func NewSDK(registry services.Registry) *SDK {
	return NewFrom(fabric.NewSDK(registry))
}

func NewFrom(sdk common.SDK) *SDK {
	return &SDK{SDK: sdk}
}

func (p *SDK) FabricEnabled() bool {
	return p.ConfigService().GetBool("fabric.enabled")
}

func (p *SDK) Install() error {
	if !p.FabricEnabled() {
		return p.SDK.Install()
	}
	err := errors.Join(
		// Register the new fabricx platform driver
		p.Container().Provide(NewDriver, dig.Group("fabric-platform-drivers")),
		p.Container().Provide(NewChannelProvider, dig.As(new(ChannelProvider))),
		p.Container().Provide(ledger.NewEventBasedProvider, dig.As(new(ledger.Provider))),
		p.Container().Provide(ledger.NewBlockDispatcherProvider),
		p.Container().Provide(finality.NewListenerManagerProvider),
		p.Container().Provide(queryservice.NewProvider),
		p.Container().Provide(namespace.NewSubmitterFromFNS, dig.As(new(namespace.Submitter))),
		p.Container().Provide(namespace.NewDeployerServiceFromFNS, dig.As(new(namespace.DeployerService))),
		p.Container().Provide(protoblocktx2.NewStaticMappingService, dig.As(new(protoblocktx2.MappingService))),
		p.Container().Provide(func(service driver.ConfigService, mapper protoblocktx2.MappingService) (protoqueryservice.Provider, error) {
			p, err := protoqueryservice.NewProvider(service)
			if err != nil {
				return nil, err
			}
			return p.
				Register(v1.CommitterVersion, protoqueryservice2.NewQueryServiceClientProvider(mapper)).
				Register(v2.CommitterVersion, protoqueryservice3.NewQueryServiceClientProvider()), nil
		}),
		p.Container().Provide(func(service driver.ConfigService, mapper protoblocktx2.MappingService) (protoblocktx.Provider, error) {
			p, err := protoblocktx.NewProvider(service)
			if err != nil {
				return nil, err
			}
			return p.
				Register(v1.CommitterVersion, protoblocktx2.NewMarshallerAdapter(mapper)).
				Register(v2.CommitterVersion, protoblocktx3.NewMarshallerAdapter()), nil
		}),
	)
	if err != nil {
		return err
	}

	err = errors.Join(
		p.Container().Decorate(NewConfigProvider),
	)
	if err != nil {
		return err
	}

	if err := p.SDK.Install(); err != nil {
		return err
	}

	// Backward compatibility with SP
	return errors.Join(
		digutils.Register[namespace.DeployerService](p.Container()),
		digutils.Register[finality.ListenerManagerProvider](p.Container()),
		digutils.Register[queryservice.Provider](p.Container()),
	)
}

func (p *SDK) Start(ctx context.Context) error {
	return p.SDK.Start(ctx)
}
