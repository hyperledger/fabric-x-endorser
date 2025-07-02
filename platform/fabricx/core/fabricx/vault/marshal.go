/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package vault

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/core/generic/vault"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/hash"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Marshaller is the custom marshaller for fabricx.
type Marshaller struct {
	adapter protoblocktx.Marshaller

	NsInfo map[driver.Namespace]driver.RawVersion
}

func NewMarshaller(adapter protoblocktx.Marshaller) *Marshaller {
	return &Marshaller{adapter: adapter}
}

func (m *Marshaller) Marshal(txID string, rws *vault.ReadWriteSet) ([]byte, error) {
	logger.Debugf("Marshal rws into fabricx proto [txID=%v]", txID)
	return m.marshal(txID, rws, m.NsInfo)
}

func (m *Marshaller) marshal(txID string, rws *vault.ReadWriteSet, nsInfo map[driver.Namespace]driver.RawVersion) ([]byte, error) {
	if logger.IsEnabledFor(zap.DebugLevel) {
		str, _ := json.MarshalIndent(rws, "", "\t")
		logger.Debugf("Marshal vault.ReadWriteSet %s", string(str))
	}

	type namespaceType struct {
		ns           driver.Namespace
		nsVersion    driver.RawVersion
		readSet      map[string]protoblocktx.Read
		writeSet     map[string]protoblocktx.Write
		readWriteSet map[string]protoblocktx.ReadWrite
	}

	newNamespace := func(ns driver.Namespace, nsVersion driver.RawVersion) *namespaceType {
		return &namespaceType{
			ns:           ns,
			nsVersion:    nsVersion,
			readSet:      make(map[string]protoblocktx.Read),
			writeSet:     make(map[string]protoblocktx.Write),
			readWriteSet: make(map[string]protoblocktx.ReadWrite),
		}
	}

	namespaceSet := make(map[driver.Namespace]*namespaceType)

	// writes ...
	for ns, keyMap := range rws.Writes {
		// check that namespace exists as in _meta
		nsVersion, exists := nsInfo[ns]
		if !exists {
			panic(fmt.Sprintf("ns = [%s] does not exist in nsInfo", ns))
		}

		// create namespace if not already exists
		namespace, exists := namespaceSet[ns]
		if !exists {
			namespace = newNamespace(ns, nsVersion)
			namespaceSet[ns] = namespace
		}

		for key, val := range keyMap {
			namespace.writeSet[key] = protoblocktx.NewWrite([]byte(key), val)
			logger.Debugf("blind write [%s:%s][%s]", namespace.ns, key, val)
		}
	}

	// reads
	for ns, keyMap := range rws.Reads {
		// check that namespace exists as in _meta
		nsVersion, exists := nsInfo[ns]
		if !exists {
			panic(fmt.Sprintf("ns = [%s] does not exist in nsInfo", ns))
		}

		// create namespace if not already exists
		namespace, exists := namespaceSet[ns]
		if !exists {
			namespace = newNamespace(ns, nsVersion)
			namespaceSet[ns] = namespace
		}

		for key, ver := range keyMap {
			// let's check if our read is a read-write or read-only
			if w, exists := namespace.writeSet[key]; exists {
				namespace.readWriteSet[key] = protoblocktx.NewReadWrite([]byte(key), ver, w.GetValue())
				logger.Debugf("blind write was a read write [%s:%s][%s][%s]", namespace.ns, key, w.GetValue(), types.VersionNumberFromBytes(ver))
				delete(namespace.writeSet, key)
			} else {
				namespace.readSet[key] = protoblocktx.NewRead([]byte(key), ver)
				logger.Debugf("read [%s:%s][%v]", namespace.ns, key, types.VersionNumberFromBytes(ver))
			}
		}
	}

	namespaces := make([]protoblocktx.TxNamespace, 0)
	for _, namespace := range namespaceSet {
		readsOnly := make([]protoblocktx.Read, 0, len(namespace.readSet))
		for _, read := range namespace.readSet {
			readsOnly = append(readsOnly, read)
		}

		blindWrites := make([]protoblocktx.Write, 0, len(namespace.writeSet))
		for _, write := range namespace.writeSet {
			blindWrites = append(blindWrites, write)
		}

		readWrites := make([]protoblocktx.ReadWrite, 0, len(namespace.readWriteSet))
		for _, readWrite := range namespace.readWriteSet {
			readWrites = append(readWrites, readWrite)
		}

		namespaces = append(namespaces, protoblocktx.NewTxNamespace(namespace.ns, namespace.nsVersion, readsOnly, readWrites, blindWrites))
	}

	txIn := protoblocktx.NewTx(txID, namespaces, nil)
	if logger.IsEnabledFor(zapcore.DebugLevel) {
		str, _ := json.MarshalIndent(txIn, "", "\t")
		logger.Debugf("fabricx transaction: %s", str)
	}

	if logger.IsEnabledFor(zap.DebugLevel) {
		str, _ := json.MarshalIndent(txIn, "", "\t")
		logger.Debugf("Unmarshalled fabricx tx %s", string(str))
	}

	raw, err := m.adapter.MarshalTx(txIn)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func (m *Marshaller) RWSetFromBytes(raw []byte, namespaces ...string) (*vault.ReadWriteSet, error) {
	rws := vault.EmptyRWSet()
	if err := m.Append(&rws, raw, namespaces...); err != nil {
		return nil, err
	}
	return &rws, nil
}

func (m *Marshaller) Append(destination *vault.ReadWriteSet, raw []byte, namespaces ...string) error {
	txIn, err := m.adapter.UnmarshalTx(raw)
	if err != nil {
		return errors.Wrapf(err, "failed unmarshalling tx from (%d)[%s]", len(raw), hash.Hashable(raw))
	}

	if logger.IsEnabledFor(zap.DebugLevel) {
		str, _ := json.MarshalIndent(txIn, "", "\t")
		logger.Debugf("Unmarshalled fabricx tx %s", string(str))
	}

	for _, txNs := range txIn.GetNamespaces() {
		for _, read := range txNs.GetReadsOnly() {
			destination.ReadSet.Add(txNs.GetNsId(), string(read.GetKey()), read.GetVersion())
		}

		for _, write := range txNs.GetBlindWrites() {
			if err := destination.WriteSet.Add(txNs.GetNsId(), string(write.GetKey()), write.GetValue()); err != nil {
				// TODO: ... should we really just stop here or revert all changes ... ?
				return errors.Wrapf(err, "failed adding blindwrite [%s]", write.GetKey())
			}
		}

		for _, readWrite := range txNs.GetReadWrites() {
			destination.ReadSet.Add(txNs.GetNsId(), string(readWrite.GetKey()), readWrite.GetVersion())
			if err := destination.WriteSet.Add(txNs.GetNsId(), string(readWrite.GetKey()), readWrite.GetValue()); err != nil {
				// TODO: ... should we really just stop here or revert all changes ... ?
				return errors.Wrapf(err, "failed adding readwrite [%s]", readWrite.GetKey())
			}
		}
	}

	return nil
}
