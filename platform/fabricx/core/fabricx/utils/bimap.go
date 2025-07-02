/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"sync"

	"github.com/hyperledger-labs/fabric-smart-client/platform/common/utils"
)

func NewBiMap[K, V comparable](update func() (map[K]V, error)) *BiMap[K, V] {
	return &BiMap[K, V]{
		byKey:   map[K]V{},
		byValue: map[V]K{},
		update:  update,
	}
}

type BiMap[K comparable, V comparable] struct {
	byKey   map[K]V
	byValue map[V]K
	mu      sync.RWMutex
	update  func() (map[K]V, error)
}

func (m *BiMap[K, V]) load() error {
	kvs, err := m.update()
	if err != nil {
		return err
	}

	for k, v := range kvs {
		m.byKey[k] = v
		m.byValue[v] = k
	}
	return nil
}

func (m *BiMap[K, V]) Get(k K) (V, bool) {
	return get(k, m.byKey, &m.mu)
}

func (m *BiMap[K, V]) InverseGet(v V) (K, bool) {
	return get(v, m.byValue, &m.mu)
}

func (m *BiMap[K, V]) GetOrUpdate(k K) (V, error) {
	return getOrUpdate(k, m.byKey, &m.mu, m.load)
}

func (m *BiMap[K, V]) InverseGetOrUpdate(v V) (K, error) {
	return getOrUpdate(v, m.byValue, &m.mu, m.load)
}

func getOrUpdate[K comparable, V any](k K, m map[K]V, mu *sync.RWMutex, load func() error) (V, error) {
	if v, ok := get(k, m, mu); ok {
		return v, nil
	}

	mu.Lock()
	defer mu.Unlock()
	if v, ok := m[k]; ok {
		return v, nil
	}

	if err := load(); err != nil {
		return utils.Zero[V](), err
	}
	return m[k], nil
}

func get[K comparable, V any](k K, m map[K]V, mu *sync.RWMutex) (V, bool) {
	mu.RLock()
	defer mu.RUnlock()
	v, ok := m[k]
	return v, ok
}
