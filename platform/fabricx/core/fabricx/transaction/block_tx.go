/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package transaction

import (
	"encoding/asn1"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/hash"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
)

func HashTxNamespace(txID string, ns protoblocktx.TxNamespace) []byte {
	bytes, err := asn1.Marshal(Tx{
		TxID:      txID,
		Namespace: hashTxNamespace(ns),
	})
	if err != nil {
		// TODO better way to deal with an error here
		panic(errors.Wrap(err, "cannot hash transaction due to marshaling error"))
	}

	h, err := hash.SHA256(bytes)
	if err != nil {
		panic(err)
	}
	return h
}

func hashTxNamespace(ns protoblocktx.TxNamespace) Namespace {
	n := Namespace{
		Reads:      make([]Reads, len(ns.GetReadsOnly())),
		ReadWrites: make([]ReadWrites, len(ns.GetReadWrites())),
		Writes:     make([]BlindWrites, len(ns.GetBlindWrites())),
	}

	for i, r := range ns.GetReadsOnly() {
		n.Reads[i] = Reads{
			Key:     r.GetKey(),
			Version: r.GetVersion(),
		}
	}

	for i, rw := range ns.GetReadWrites() {
		n.ReadWrites[i] = ReadWrites{
			Key:     rw.GetKey(),
			Version: rw.GetVersion(),
			Value:   rw.GetValue(),
		}
	}

	for i, bw := range ns.GetBlindWrites() {
		n.Writes[i] = BlindWrites{
			Key:     bw.GetKey(),
			Version: bw.GetValue(),
			Value:   nil,
		}
	}
	return n
}

type Tx struct {
	TxID      string
	Namespace Namespace
}

type Namespace struct {
	Reads      []Reads
	ReadWrites []ReadWrites
	Writes     []BlindWrites
}

type Reads struct {
	Key     []byte
	Version []byte
}

type BlindWrites struct {
	Key     []byte
	Version []byte
	Value   []byte
}

type ReadWrites struct {
	Key     []byte
	Version []byte
	Value   []byte
}
