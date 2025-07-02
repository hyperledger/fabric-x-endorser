/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoblocktx

import (
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/types"
)

const MetaNamespace driver.Namespace = "_meta"

type Provider = types.Provider[Marshaller]

func NewProvider(configService driver.ConfigService) (*types.ServiceProvider[Marshaller], error) {
	return types.NewProvider[Marshaller](configService)
}

type Marshaller interface {
	UnmarshalTx([]byte) (Tx, error)
	MarshalTx(Tx) ([]byte, error)
	MarshalNamespacePolicy(NamespacePolicy) ([]byte, error)
	MarshalNamespaceID(driver.Namespace) ([]byte, error)
	IsStatusValid(b byte) bool
}
