/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoblocktx

import (
	"regexp"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/proto"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	api "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/utils"
	"github.com/hyperledger/fabric/protoutil"
)

const metaNamespaceId = "_meta"

type marshallerAdapter struct{}

func NewMarshallerAdapter() *marshallerAdapter {
	return &marshallerAdapter{}
}

func (a *marshallerAdapter) UnmarshalTx(raw []byte) (api.Tx, error) {
	var tx Tx
	if err := proto.Unmarshal(raw, &tx); err != nil {
		return nil, err
	}
	namespaces := make([]api.TxNamespace, len(tx.GetNamespaces()))
	for i, ns := range tx.GetNamespaces() {
		nsID := ns.GetNsId()
		if nsID == metaNamespaceId {
			nsID = api.MetaNamespace
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
	return protoutil.Marshal(&Tx{
		Id:         tx.GetId(),
		Namespaces: utils.Map(tx.GetNamespaces(), mapTxNamespace),
		Signatures: tx.GetSignatures(),
	})
}

func (a *marshallerAdapter) MarshalNamespacePolicy(p api.NamespacePolicy) ([]byte, error) {
	return protoutil.Marshal(&NamespacePolicy{Scheme: p.GetScheme(), PublicKey: p.GetPublicKey()})
}

func (a *marshallerAdapter) MarshalNamespaceID(nsID driver.Namespace) ([]byte, error) {
	if err := validateNamespaceID(nsID); err != nil {
		return nil, err
	}
	return []byte(nsID), nil
}

func (a *marshallerAdapter) IsStatusValid(b byte) bool {
	return b == byte(Status_COMMITTED)
}

// maxNamespaceIDLength defines the maximum number of characters allowed for namespace IDs.
// PostgreSQL limits identifiers to NAMEDATALEN-1, where NAMEDATALEL=64.
// The namespace tables have the prefix 'ns_', thus there are 60 characters remaining.
// See: https://www.postgresql.org/docs/current/sql-syntax-lexical.html
const maxNamespaceIDLength = 60

// validNamespaceID describes the allowed characters in a namespace ID.
// The name may contain letters, digits, or underscores.
// PostgreSQL requires the name to begin with a letter. This is ensured by our namespace table prefix.
// In addition, we restrict to lowercase letters as PostgreSQL converts does not distinguish between upper/lower case.
// See: https://www.postgresql.org/docs/current/sql-syntax-lexical.html
// The regexp is wrapped with "^...$" to ensure we only match the entire string (namespace ID).
// We use the flags: i - ignore case, s - single line.
var validNamespaceID = regexp.MustCompile(`^[a-z0-9_]+$`)

// ErrInvalidNamespaceID is returned when the namespace ID cannot be parsed.
var ErrInvalidNamespaceID = errors.New("invalid namespace ID")

func validateNamespaceID(nsID driver.Namespace) error {
	// if it matches our holy MetaNamespaceID it is valid.
	if nsID == api.MetaNamespace {
		return nil
	}

	// length checks.
	if len(nsID) == 0 || len(nsID) > maxNamespaceIDLength {
		return ErrInvalidNamespaceID
	}

	// characters check.
	if !validNamespaceID.MatchString(nsID) {
		return ErrInvalidNamespaceID
	}

	return nil
}

func mapTxNamespace(ns api.TxNamespace) *TxNamespace {
	if ns == nil {
		return nil
	}
	return &TxNamespace{
		NsId:        mapNsId(ns.GetNsId()),
		NsVersion:   ns.GetNsVersion(),
		ReadsOnly:   utils.Map(ns.GetReadsOnly(), mapRead),
		ReadWrites:  utils.Map(ns.GetReadWrites(), mapReadWrite),
		BlindWrites: utils.Map(ns.GetBlindWrites(), mapWrite),
	}
}

func mapNsId(nsId string) driver.Namespace {
	if nsId == api.MetaNamespace {
		return metaNamespaceId
	}
	return nsId
}

func mapRead(r api.Read) *Read {
	if r == nil {
		return nil
	}
	return &Read{Key: r.GetKey(), Version: r.GetVersion()}
}

func mapWrite(w api.Write) *Write {
	if w == nil {
		return nil
	}
	return &Write{Key: w.GetKey(), Value: w.GetValue()}
}

func mapReadWrite(rw api.ReadWrite) *ReadWrite {
	if rw == nil {
		return nil
	}
	return &ReadWrite{Key: rw.GetKey(), Version: rw.GetVersion(), Value: rw.GetValue()}
}
