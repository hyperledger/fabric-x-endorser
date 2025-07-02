/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package namespace

import (
	"time"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/fabricutils"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/transaction"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	transaction2 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/transaction"
	"github.com/hyperledger/fabric/protoutil"
)

type Submitter interface {
	Submit(network, channel string, namespaces []protoblocktx.TxNamespace) error
}

const (
	finalityEOFRetries    = 5
	finalityRetryDuration = 2 * time.Second
)

func NewSubmitterFromFNS(fnsp *fabric.NetworkServiceProvider, adapterProvider protoblocktx.Provider) *submitter {
	return NewSubmitter(&fnsSigningIdentityProvider{fnsProvider: fnsp}, &fnsBroadcaster{fnsProvider: fnsp}, adapterProvider)
}

func NewSubmitter(signingIdentityProvider SigningIdentityProvider, envelopeBroadcaster EnvelopeBroadcaster, adapterProvider protoblocktx.Provider) *submitter {
	return NewSubmitterCustomTxID(signingIdentityProvider, envelopeBroadcaster, adapterProvider, protoutil.ComputeTxID)
}

func NewSubmitterCustomTxID(signingIdentityProvider SigningIdentityProvider, envelopeBroadcaster EnvelopeBroadcaster, adapterProvider protoblocktx.Provider, txIDCalculator func(nonce, creator []byte) string) *submitter {
	return &submitter{
		txIDCalculator:          txIDCalculator,
		signingIdentityProvider: signingIdentityProvider,
		envelopeBroadcaster:     envelopeBroadcaster,
		adapterProvider:         adapterProvider,
	}
}

type submitter struct {
	txIDCalculator          func(nonce, creator []byte) string
	signingIdentityProvider SigningIdentityProvider
	envelopeBroadcaster     EnvelopeBroadcaster
	adapterProvider         protoblocktx.Provider
}

func (s *submitter) Submit(network, channel string, namespaces []protoblocktx.TxNamespace) error {
	logger.Infof("Submitting to [%s,%s] following %d namespaces: [%v]", network, channel, len(namespaces), namespaces)

	signer, err := s.signingIdentityProvider.DefaultSigningIdentity(network, channel)
	if err != nil {
		return err
	}

	serializedCreator, err := s.signingIdentityProvider.DefaultIdentity(network, channel)
	if err != nil {
		return err
	}

	nonce, err := transaction.GetRandomNonce()
	if err != nil {
		return errors.Wrapf(err, "failed getting random nonce")
	}

	txID := s.txIDCalculator(nonce, serializedCreator)

	// compute namespace tx hash
	sigs := make([][]byte, len(namespaces))
	for i, namespace := range namespaces {
		sigs[i], err = signer.Sign(transaction2.HashTxNamespace(txID, namespace))
		if err != nil {
			return errors.Wrapf(err, "failed signing tx")
		}
	}

	nsTx := protoblocktx.NewTx(txID, namespaces, sigs)

	adapter, err := s.adapterProvider.Get(network, channel)
	if err != nil {
		return err
	}
	txRaw, err := adapter.MarshalTx(nsTx)
	if err != nil {
		return errors.Wrapf(err, "failed marshaling transaction")
	}

	signatureHeader := &common.SignatureHeader{Creator: serializedCreator, Nonce: nonce}
	channelHeader := protoutil.MakeChannelHeader(common.HeaderType_MESSAGE, 0, channel, 0)
	channelHeader.TxId = txID
	payloadHeader := protoutil.MakePayloadHeader(channelHeader, signatureHeader)
	env, err := fabricutils.CreateEnvelope(signer, payloadHeader, txRaw)
	if err != nil {
		return errors.Wrapf(err, "failed creating envelope")
	}

	return s.envelopeBroadcaster.Broadcast(network, channel, txID, env)
}
