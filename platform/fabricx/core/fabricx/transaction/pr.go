/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package transaction

import (
	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/proto"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
)

type VerifierProvider = driver.VerifierProvider

type ProposalResponse struct {
	pr      *peer.ProposalResponse
	results []byte
	adapter protoblocktx.Marshaller
}

func NewProposalResponseFromResponse(proposalResponse *peer.ProposalResponse, adapter protoblocktx.Marshaller) (*ProposalResponse, error) {
	return &ProposalResponse{
		pr:      proposalResponse,
		results: proposalResponse.Payload,
		adapter: adapter,
	}, nil
}

func NewProposalResponseFromBytes(raw []byte, adapter protoblocktx.Marshaller) (*ProposalResponse, error) {
	proposalResponse := &peer.ProposalResponse{}
	if err := proto.Unmarshal(raw, proposalResponse); err != nil {
		return nil, errors.Wrap(err, "failed unmarshalling received proposal response")
	}
	return NewProposalResponseFromResponse(proposalResponse, adapter)
}

func (p *ProposalResponse) Endorser() []byte {
	return p.pr.Endorsement.Endorser
}

func (p *ProposalResponse) Payload() []byte {
	return p.pr.Payload
}

func (p *ProposalResponse) EndorserSignature() []byte {
	return p.pr.Endorsement.Signature
}

func (p *ProposalResponse) Results() []byte {
	return p.results
}

func (p *ProposalResponse) PR() *peer.ProposalResponse {
	return p.pr
}

func (p *ProposalResponse) ResponseStatus() int32 {
	return p.pr.Response.Status
}

func (p *ProposalResponse) ResponseMessage() string {
	return p.pr.Response.Message
}

func (p *ProposalResponse) Bytes() ([]byte, error) {
	raw, err := proto.Marshal(p.pr)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (p *ProposalResponse) VerifyEndorsement(provider VerifierProvider) error {
	endorser := view.Identity(p.pr.Endorsement.Endorser)
	v, err := provider.GetVerifier(endorser)
	if err != nil {
		return errors.Wrapf(err, "failed getting verifier for [%s]", endorser)
	}
	// unmarshal payload to Tx
	tx, err := p.adapter.UnmarshalTx(p.pr.Payload)
	if err != nil {
		return errors.Wrapf(err, "failed unmarshalling payload for [%s]", endorser)
	}
	msg := HashTxNamespace(tx.GetId(), tx.GetNamespaces()[0])
	return v.Verify(msg, p.EndorserSignature())
}
