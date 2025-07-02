/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoblocktx

import (
	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/proto"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	api "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/utils"
	"google.golang.org/protobuf/encoding/protowire"
)

type MappingService interface {
	IDByName(name string) (uint32, error)
	NameByID(id uint32) (string, error)
}

type marshallerAdapter struct {
	nsMapper MappingService
}

func NewMarshallerAdapter(nsMapper MappingService) *marshallerAdapter {
	return &marshallerAdapter{nsMapper: nsMapper}
}

func (a *marshallerAdapter) UnmarshalTx(raw []byte) (api.Tx, error) {
	var tx Tx
	if err := proto.Unmarshal(raw, &tx); err != nil {
		return nil, err
	}

	namespaces := make([]api.TxNamespace, len(tx.GetNamespaces()))
	for i, ns := range tx.GetNamespaces() {
		nsID, err := a.nsMapper.NameByID(ns.GetNsId())
		if err != nil {
			return nil, err
		}
		namespaces[i] = api.NewTxNamespace(
			nsID,
			ns.GetNsVersion(),
			utils.Map(ns.GetReadsOnly(), func(r *Read) api.Read { return r }),
			utils.Map(ns.GetReadWrites(), func(rw *ReadWrite) api.ReadWrite { return rw }),
			utils.Map(ns.GetBlindWrites(), func(w *Write) api.Write { return w }))
	}
	return api.NewTx(tx.GetId(), namespaces, tx.GetSignatures()), nil
}

func (a *marshallerAdapter) MarshalTx(tx api.Tx) ([]byte, error) {
	if tx == nil {
		return nil, nil
	}
	namespaces := make([]*TxNamespace, len(tx.GetNamespaces()))
	for i, ns := range tx.GetNamespaces() {
		nsID, err := a.nsMapper.IDByName(ns.GetNsId())
		if err != nil {
			return nil, err
		}
		namespaces[i] = &TxNamespace{
			NsId:        nsID,
			NsVersion:   ns.GetNsVersion(),
			ReadsOnly:   utils.Map(ns.GetReadsOnly(), mapRead),
			ReadWrites:  utils.Map(ns.GetReadWrites(), mapReadWrite),
			BlindWrites: utils.Map(ns.GetBlindWrites(), mapWrite),
		}
	}
	return proto.Marshal(&Tx{
		Id:         tx.GetId(),
		Namespaces: namespaces,
		Signatures: tx.GetSignatures(),
	})
}

func (a *marshallerAdapter) MarshalNamespacePolicy(p api.NamespacePolicy) ([]byte, error) {
	return proto.Marshal(&NamespacePolicy{Scheme: p.GetScheme(), PublicKey: p.GetPublicKey()})
}

func (a *marshallerAdapter) MarshalNamespaceID(ns driver.Namespace) ([]byte, error) {
	nsID, err := a.nsMapper.IDByName(ns)
	if err != nil {
		return nil, err
	}
	// To read a namespace:
	//	v, l := protowire.ConsumeVarint(ns)
	//	if l < 0 || l != len(ns) {
	//		return 0, fmt.Errorf("invalid namespace id")
	//	}
	return protowire.AppendVarint(nil, uint64(nsID)), nil
}

func (a *marshallerAdapter) IsStatusValid(b byte) bool { return b == byte(Status_COMMITTED) }

func mapRead(r api.Read) *Read { return &Read{Key: r.GetKey(), Version: r.GetVersion()} }

func mapWrite(w api.Write) *Write { return &Write{Key: w.GetKey(), Value: w.GetValue()} }

func mapReadWrite(rw api.ReadWrite) *ReadWrite {
	return &ReadWrite{Key: rw.GetKey(), Version: rw.GetVersion(), Value: rw.GetValue()}
}
