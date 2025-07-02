/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package namespace

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/msp/x509"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/ordering"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/services"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/grpc"
	view2 "github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/types"
	v1 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v1"
	protoblocktx2 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v1/protoblocktx"
	v2 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v2"
	protoblocktx3 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v2/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/namespace"
)

var marshallers = map[types.CommitterVersion]protoblocktx.Marshaller{
	// TODO: v1 should be removed; in particular, the static mapping hack here
	v1.CommitterVersion: protoblocktx2.NewMarshallerAdapter(protoblocktx2.NewStaticMappingService(protoblocktx2.Namespace{ID: 1, Name: "iou"}, protoblocktx2.Namespace{ID: 2, Name: "boring"})),
	v2.CommitterVersion: protoblocktx3.NewMarshallerAdapter(),
}

type OrdererConfig struct {
	OrderingEndpoint           string
	TLSEnabled                 bool
	ClientAuth                 bool
	CaFile                     string
	KeyFile                    string
	CertFile                   string
	OrdererTLSHostnameOverride string
	ConnTimeout                time.Duration
	TLSHandshakeTimeShift      time.Duration
}

type MSPConfig struct {
	MSPConfigPath string
	MSPID         string
}

type dummyAdapterProvider struct{ m protoblocktx.Marshaller }

func (p *dummyAdapterProvider) Get(network, channel string) (protoblocktx.Marshaller, error) {
	return p.m, nil
}

func DeployNamespace(chName, nsID string, nsVersion int, committerVersion types.CommitterVersion, odererCfg OrdererConfig, mspCfg MSPConfig, pkPath string) error {
	sip := &signingIdentityProvider{mspCfg: mspCfg}
	sid, err := sip.DefaultSigningIdentity("", "")
	if err != nil {
		return err
	}
	bp := &broadcaster{odererCfg: odererCfg, signer: sid}
	adapter, ok := marshallers[committerVersion]
	if !ok {
		return fmt.Errorf("version %s not found", committerVersion)
	}
	adapterProvider := &dummyAdapterProvider{m: adapter}

	submitter := namespace.NewSubmitter(sip, bp, adapterProvider)
	deployer := namespace.NewDeployerService(adapterProvider, submitter, sip)

	return deployer.DeployNamespaceWithKeyAndVersion("", chName, nsID, nsVersion, pkPath)
}

type signingIdentityProvider struct {
	mspCfg MSPConfig
}

func (p *signingIdentityProvider) DefaultSigningIdentity(network, channel string) (namespace.Signer, error) {
	return x509.GetSigningIdentity(p.mspCfg.MSPConfigPath, "", p.mspCfg.MSPID, nil)
}

func (p *signingIdentityProvider) DefaultIdentity(network, channel string) (view2.Identity, error) {
	si, err := p.DefaultSigningIdentity(network, channel)
	if err != nil {
		return nil, err
	}
	return si.Serialize()
}

type broadcaster struct {
	odererCfg OrdererConfig
	signer    driver.Signer
}

func (b *broadcaster) Broadcast(network, channel string, txID driver.TxID, env *common.Envelope) error {
	secOpts, err := grpc.CreateSecOpts(grpc.ConnectionConfig{
		Address:            b.odererCfg.OrderingEndpoint,
		ConnectionTimeout:  b.odererCfg.ConnTimeout,
		TLSEnabled:         b.odererCfg.TLSEnabled,
		TLSClientSideAuth:  b.odererCfg.ClientAuth,
		TLSDisabled:        !b.odererCfg.TLSEnabled,
		ServerNameOverride: b.odererCfg.OrdererTLSHostnameOverride,
		TLSRootCertFile:    b.odererCfg.CaFile,
	}, grpc.TLSClientConfig{
		TLSClientAuthRequired: b.odererCfg.ClientAuth,
		TLSClientKeyFile:      b.odererCfg.KeyFile,
		TLSClientCertFile:     b.odererCfg.CertFile,
	})
	if err != nil {
		return err
	}

	gClient, err := grpc.NewGRPCClient(grpc.ClientConfig{
		SecOpts: *secOpts,
		Timeout: b.odererCfg.ConnTimeout,
	})
	if err != nil {
		return err
	}

	orderingClient := services.NewGRPCClient(gClient, b.odererCfg.OrderingEndpoint, b.odererCfg.OrdererTLSHostnameOverride, b.signer.Sign)

	occ, err := orderingClient.OrdererClient()
	if err != nil {
		return err
	}

	abc, err := occ.Broadcast(context.TODO())
	if err != nil {
		return err
	}

	conn := ordering.Connection{
		Stream: abc,
		Client: orderingClient,
	}

	err = conn.Send(env)
	if err != nil {
		return err
	}

	status, err := conn.Recv()
	if err != nil {
		return err
	}

	if status.GetStatus() != common.Status_SUCCESS {
		return fmt.Errorf("got error %#v", status.GetStatus())
	}

	return nil
}
