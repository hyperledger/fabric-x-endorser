/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package committer

import (
	"context"
	"fmt"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/services/logging"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/committer"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/driver"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

var logger = logging.MustGetLogger("fabricx.committer")

func RegisterTransactionHandler(com *committer.Committer) {
	h := NewHandler(com)

	com.Handlers[common.HeaderType_MESSAGE] = h.HandleFabricxTransaction
}

func NewHandler(com *committer.Committer) *handler {
	return &handler{committer: com}
}

type handler struct {
	committer *committer.Committer
}

func (h *handler) HandleFabricxTransaction(ctx context.Context, block *common.BlockMetadata, tx committer.CommitTx) (*committer.FinalityEvent, error) {
	logger.Debugf("Handle a new fabricx transaction [channel=%s] with [txID=%s]", h.committer.ChannelConfig.ID(), tx.TxID)

	if len(block.Metadata) < int(common.BlockMetadataIndex_TRANSACTIONS_FILTER) {
		return nil, fmt.Errorf("block metadata lacks transaction filter")
	}

	fabricValidationCode := committer.ValidationFlags(block.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])[tx.TxNum]
	event := &committer.FinalityEvent{
		Ctx:               ctx,
		TxID:              tx.TxID,
		ValidationCode:    convertValidationCode(int32(fabricValidationCode)),
		ValidationMessage: peer.TxValidationCode_name[int32(fabricValidationCode)],
	}

	switch peer.TxValidationCode(fabricValidationCode) {
	case peer.TxValidationCode_VALID:
		processed, err := h.committer.CommitEndorserTransaction(ctx, event.TxID, tx.BlkNum, tx.TxNum, tx.Envelope, event)
		if err != nil {
			if errors.HasCause(err, committer.ErrDiscardTX) {
				// in this case, we will discard the transaction
				event.ValidationCode = driver.Invalid
				event.ValidationMessage = err.Error()

				// escaping the switch and discard
				break
			}
			return nil, fmt.Errorf("failed committing transaction [%s]: %w", event.TxID, err)
		}
		if !processed {
			logger.Debugf("TODO: Should we try to get chaincode events?")
			//if err := h.committer.GetChaincodeEvents(tx.Envelope, tx.BlkNum); err != nil {
			//	return nil, fmt.Errorf("failed to publish chaincode events [%s]: %w", event.TxID, err)
			//}
		}
		return event, nil
	}

	logger.Warnf("Discarding transaction %s", tx.TxID)
	if err := h.committer.DiscardEndorserTransaction(ctx, event.TxID, tx.BlkNum, tx.Raw, event); err != nil {
		return nil, fmt.Errorf("failed discarding transaction [%s]: %w", event.TxID, err)
	}

	return event, nil
}

func convertValidationCode(vc int32) driver.ValidationCode {
	switch peer.TxValidationCode(vc) {
	case peer.TxValidationCode_VALID:
		return driver.Valid
	default:
		return driver.Invalid
	}
}
