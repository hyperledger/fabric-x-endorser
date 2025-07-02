/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoblocktx

import (
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/utils"
)

const metaNamespaceId uint32 = 1024

type Namespaces []Namespace

type Namespace struct {
	ID   uint32
	Name string
}

func (c Namespaces) AsMap() map[uint32]string {
	m := make(map[uint32]string, len(c))
	for _, ns := range c {
		m[ns.ID] = ns.Name
	}
	return m
}

func NewStaticMappingService(nss ...Namespace) *staticService {
	namespaces := Namespaces(append(nss, Namespace{ID: metaNamespaceId, Name: protoblocktx.MetaNamespace}))
	return &staticService{
		nss: utils.NewBiMap(func() (map[uint32]string, error) { return namespaces.AsMap(), nil }),
	}
}

type staticService struct {
	nss *utils.BiMap[uint32, string]
}

func (m *staticService) IDByName(name string) (uint32, error) {
	return m.nss.InverseGetOrUpdate(name)
}

func (m *staticService) NameByID(id uint32) (string, error) { return m.nss.GetOrUpdate(id) }
