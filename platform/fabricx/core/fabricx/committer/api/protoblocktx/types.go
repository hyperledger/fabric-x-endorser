/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoblocktx

import (
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
)

type tx struct {
	Id         string
	Namespaces []TxNamespace
	Signatures [][]byte
}

func NewTx(Id string, Namespaces []TxNamespace, Signatures [][]byte) *tx {
	return &tx{
		Id:         Id,
		Namespaces: Namespaces,
		Signatures: Signatures,
	}
}

func (t *tx) GetId() string                { return t.Id }
func (t *tx) GetNamespaces() []TxNamespace { return t.Namespaces }
func (t *tx) GetSignatures() [][]byte      { return t.Signatures }

type Tx interface {
	GetId() string
	GetNamespaces() []TxNamespace
	GetSignatures() [][]byte
}

type txNamespace struct {
	NsId        driver.Namespace
	NsVersion   []byte
	ReadsOnly   []Read
	ReadWrites  []ReadWrite
	BlindWrites []Write
}

func NewTxNamespace(NsId driver.Namespace, NsVersion []byte, ReadsOnly []Read, ReadWrites []ReadWrite, BlindWrites []Write) *txNamespace {
	return &txNamespace{
		NsId:        NsId,
		NsVersion:   NsVersion,
		ReadsOnly:   ReadsOnly,
		ReadWrites:  ReadWrites,
		BlindWrites: BlindWrites,
	}
}

func (n *txNamespace) GetNsId() driver.Namespace  { return n.NsId }
func (n *txNamespace) GetNsVersion() []byte       { return n.NsVersion }
func (n *txNamespace) GetReadsOnly() []Read       { return n.ReadsOnly }
func (n *txNamespace) GetReadWrites() []ReadWrite { return n.ReadWrites }
func (n *txNamespace) GetBlindWrites() []Write    { return n.BlindWrites }

type TxNamespace interface {
	GetNsId() driver.Namespace
	GetNsVersion() []byte
	GetReadsOnly() []Read
	GetReadWrites() []ReadWrite
	GetBlindWrites() []Write
}

type read struct {
	Key     []byte
	Version []byte
}

func NewRead(Key, Version []byte) *read {
	return &read{
		Key:     Key,
		Version: Version,
	}
}

func (r *read) GetKey() []byte     { return r.Key }
func (r *read) GetVersion() []byte { return r.Version }

type Read interface {
	GetKey() []byte
	GetVersion() []byte
}

type readWrite struct {
	Key     []byte
	Version []byte
	Value   []byte
}

func NewReadWrite(Key, Version, Value []byte) *readWrite {
	return &readWrite{
		Key:     Key,
		Version: Version,
		Value:   Value,
	}
}

func (w *readWrite) GetKey() []byte     { return w.Key }
func (w *readWrite) GetVersion() []byte { return w.Version }
func (w *readWrite) GetValue() []byte   { return w.Value }

type ReadWrite interface {
	GetKey() []byte
	GetVersion() []byte
	GetValue() []byte
}

type write struct {
	Key   []byte
	Value []byte
}

func NewWrite(Key, Value []byte) *write {
	return &write{
		Key:   Key,
		Value: Value,
	}
}

func (w *write) GetKey() []byte   { return w.Key }
func (w *write) GetValue() []byte { return w.Value }

type Write interface {
	GetKey() []byte
	GetValue() []byte
}

type namespacePolicy struct {
	Scheme    string
	PublicKey []byte
}

func NewNamespacePolicy(Scheme string, PublicKey []byte) *namespacePolicy {
	return &namespacePolicy{
		Scheme:    Scheme,
		PublicKey: PublicKey,
	}
}

func (p *namespacePolicy) GetScheme() string    { return p.Scheme }
func (p *namespacePolicy) GetPublicKey() []byte { return p.PublicKey }

type NamespacePolicy interface {
	GetScheme() string
	GetPublicKey() []byte
}
