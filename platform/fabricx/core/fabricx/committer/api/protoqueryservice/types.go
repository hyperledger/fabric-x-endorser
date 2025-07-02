/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoqueryservice

import (
	"time"

	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
)

type query struct {
	View       View
	Namespaces []QueryNamespace
}

func NewQuery(View View, Namespaces []QueryNamespace) *query {
	return &query{
		View:       View,
		Namespaces: Namespaces,
	}
}

func (q *query) GetView() View                   { return q.View }
func (q *query) GetNamespaces() []QueryNamespace { return q.Namespaces }

type Query interface {
	GetView() View
	GetNamespaces() []QueryNamespace
}

type view struct {
	Id string
}

func NewView(Id string) *view {
	return &view{Id: Id}
}

type View interface {
	GetId() string
}

type rows struct {
	Namespaces []RowsNamespace
}

func NewRows(Namespaces []RowsNamespace) *rows {
	return &rows{Namespaces: Namespaces}
}

func (r *rows) GetNamespaces() []RowsNamespace { return r.Namespaces }

type Rows interface {
	GetNamespaces() []RowsNamespace
}

type queryNamespace struct {
	NsId driver.Namespace
	Keys [][]byte
}

func NewQueryNamespace(NsId driver.Namespace, Keys [][]byte) *queryNamespace {
	return &queryNamespace{
		NsId: NsId,
		Keys: Keys,
	}
}

func (n *queryNamespace) GetNsId() driver.Namespace { return n.NsId }
func (n *queryNamespace) GetKeys() [][]byte         { return n.Keys }

type QueryNamespace interface {
	GetNsId() driver.Namespace
	GetKeys() [][]byte
}

type IsoLevel int32

type viewParameters struct {
	IsoLevel      IsoLevel
	NonDeferrable bool
	Timeout       time.Duration
}

func NewViewParameters(IsoLevel IsoLevel, NonDeferrable bool, Timeout time.Duration) *viewParameters {
	return &viewParameters{
		IsoLevel:      IsoLevel,
		NonDeferrable: NonDeferrable,
		Timeout:       Timeout,
	}
}

type ViewParameters interface {
	GetIsoLevel() IsoLevel
	GetNonDeferrable() bool
	GetTimeout() time.Duration
}

type rowsNamespace struct {
	NsId driver.Namespace
	Rows []Row
}

func NewRowsNamespace(NsId driver.Namespace, Rows []Row) *rowsNamespace {
	return &rowsNamespace{
		NsId: NsId,
		Rows: Rows,
	}
}

func (n *rowsNamespace) GetNsId() driver.Namespace { return n.NsId }
func (n *rowsNamespace) GetRows() []Row            { return n.Rows }

type RowsNamespace interface {
	GetNsId() driver.Namespace
	GetRows() []Row
}

type row struct {
	Key     []byte
	Value   []byte
	Version []byte
}

func NewRow(Key, Value, Version []byte) *row {
	return &row{
		Key:     Key,
		Value:   Value,
		Version: Version,
	}
}

type Row interface {
	GetKey() []byte
	GetValue() []byte
	GetVersion() []byte
}

type policies struct {
	Policies []PolicyItem
}

func NewPolicies(Policies []PolicyItem) *policies {
	return &policies{Policies: Policies}
}

func (p *policies) GetPolicies() []PolicyItem { return p.Policies }

type Policies interface {
	GetPolicies() []PolicyItem
}

type policyItem struct {
	Namespace string
	Policy    []byte
	Version   []byte
}

func NewPolicyItem(Namespace string, Policy, Version []byte) *policyItem {
	return &policyItem{
		Namespace: Namespace,
		Policy:    Policy,
		Version:   Version,
	}
}

type PolicyItem interface {
	GetNamespace() string
	GetPolicy() []byte
	GetVersion() []byte
}
