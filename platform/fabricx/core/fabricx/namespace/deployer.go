/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package namespace

import (
	"os"
	"reflect"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/proto"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/services/logging"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/msp/x509"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/types"
)

var logger = logging.MustGetLogger("fabricx-namespace")

type DeployerService interface {
	DeployNamespace(network, channel string, namespace driver.Namespace) error
}

func NewDeployerServiceFromFNS(
	adapterProvider protoblocktx.Provider,
	submitter Submitter,
	fnsProvider *fabric.NetworkServiceProvider,
) *deployerService {
	return &deployerService{
		adapterProvider:         adapterProvider,
		submitter:               submitter,
		signingIdentityProvider: &fnsSigningIdentityProvider{fnsProvider: fnsProvider},
	}
}

func NewDeployerService(
	adapterProvider protoblocktx.Provider,
	submitter Submitter,
	signingIdentityProvider SigningIdentityProvider,
) *deployerService {
	return &deployerService{
		adapterProvider:         adapterProvider,
		submitter:               submitter,
		signingIdentityProvider: signingIdentityProvider,
	}
}

type deployerService struct {
	signingIdentityProvider SigningIdentityProvider
	submitter               Submitter
	adapterProvider         protoblocktx.Provider
}

func (s *deployerService) DeployNamespace(network, channel, namespace string) error {
	return s.DeployNamespaceWithKeyAndVersion(network, channel, namespace, 0, "")
}

func (s *deployerService) DeployNamespaceWithKeyAndVersion(network, channel, namespace string, version int, pkPath string) error {
	var serializedPublicKey []byte

	// if `pkPath` isn't set, use the default MSP signer
	if pkPath == "" {
		sid, err := s.signingIdentityProvider.DefaultSigningIdentity(network, channel)
		if err != nil {
			return err
		}

		serializedCert, err := sid.Serialize()
		if err != nil {
			return err
		}
		serializedPublicKey, err = extractKey(serializedCert)
		if err != nil {
			return err
		}
	} else {
		b, err := os.ReadFile(pkPath)
		if err != nil {
			return err
		}

		publicKey, err := x509.PemDecodeKey(b)
		if err != nil {
			return err
		}

		serializedPublicKey, err = x509.PemEncodeKey(publicKey)
		if err != nil {
			return err
		}
	}

	adapter, err := s.adapterProvider.Get(network, channel)
	if err != nil {
		return err
	}
	namespaces, err := s.createNamespacesTxs(adapter, "ECDSA", serializedPublicKey, namespace, version)
	if err != nil {
		return err
	}
	return s.submitter.Submit(network, channel, namespaces)
}

func (s *deployerService) createNamespacesTxs(adapter protoblocktx.Marshaller, policyScheme string, policyVerificationKey []byte, nsID driver.Namespace, version int) ([]protoblocktx.TxNamespace, error) {
	policyBytes, err := adapter.MarshalNamespacePolicy(protoblocktx.NewNamespacePolicy(policyScheme, policyVerificationKey))
	if err != nil {
		return nil, err
	}

	// the version we write has to be the current version.
	// If 0, the current version is `nil`; if `n`, the current
	// version is types.VersionNumber(n-1).Bytes() etc...
	var v []byte
	if version > 0 {
		v = types.VersionNumber(version - 1).Bytes() //nolint:gosec
	}

	nsIDBytes, err := adapter.MarshalNamespaceID(nsID)
	if err != nil {
		return nil, err
	}
	return []protoblocktx.TxNamespace{protoblocktx.NewTxNamespace(
		protoblocktx.MetaNamespace,
		types.VersionNumber(0).Bytes(),
		nil,
		[]protoblocktx.ReadWrite{protoblocktx.NewReadWrite(nsIDBytes, v, policyBytes)},
		nil,
	)}, nil
}

func extractKey(serializedPublicKey []byte) ([]byte, error) {
	id := msp.SerializedIdentity{}
	if err := proto.Unmarshal(serializedPublicKey, &id); err != nil {
		return nil, err
	}
	publicKey, err := x509.PemDecodeKey(id.IdBytes)
	if err != nil {
		return nil, err
	}
	return x509.PemEncodeKey(publicKey)
}

func GetDeployerService(sp services.Provider) (DeployerService, error) {
	s, err := sp.GetService(reflect.TypeOf((*DeployerService)(nil)))
	if err != nil {
		return nil, err
	}
	return s.(DeployerService), nil
}
