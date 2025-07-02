/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoblocktx

import (
	"testing"

	api "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/stretchr/testify/require"
)

func TestMarshal(t *testing.T) {
	adapter := NewMarshallerAdapter()
	_, err := adapter.MarshalTx(api.NewTx("", []api.TxNamespace{api.NewTxNamespace("iou", []byte{0}, nil, nil, []api.Write{api.NewWrite([]byte("key"), []byte("val"))})}, nil))
	require.NoError(t, err)
}

func TestNamespaceBadNames(t *testing.T) {
	err := validateNamespaceID("go0d")
	require.NoError(t, err)

	err = validateNamespaceID("_also_good")
	require.NoError(t, err)

	err = validateNamespaceID("-bad")
	require.Error(t, err)

	err = validateNamespaceID("bad!")
	require.Error(t, err)

	err = validateNamespaceID("BAD")
	require.Error(t, err)

	err = validateNamespaceID(" bad")
	require.Error(t, err)

	err = validateNamespaceID("badbadbadbadbadbadtoolongtoolongtoolongbadbadbadbadbadbadbad1")
	require.Error(t, err)

	err = validateNamespaceID("")
	require.Error(t, err)
}
