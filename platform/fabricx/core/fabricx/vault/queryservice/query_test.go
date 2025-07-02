/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package queryservice_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/types"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v2/protoqueryservice"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v2/protoqueryservice/protoqueryservicefakes"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/vault/queryservice"
	"github.com/stretchr/testify/require"
)

func setupTest(tb testing.TB) (*queryservice.RemoteQueryService, *protoqueryservicefakes.FakeQueryServiceClient) {
	config := &queryservice.Config{
		Endpoints:    nil,
		QueryTimeout: 5 * time.Second,
	}

	client := &protoqueryservicefakes.FakeQueryServiceClient{}
	qs := queryservice.NewRemoteQueryService(config, protoqueryservice.NewServiceAdapter(client))

	return qs, client
}

func TestQueryService(t *testing.T) {
	t.Run("GetState happy path", func(t *testing.T) {
		t.Parallel()
		qs, fake := setupTest(t)

		table := []struct {
			ns       string
			key      string
			q        *protoqueryservice.Rows
			expected *driver.VaultValue
		}{
			{
				ns:  "ns1",
				key: "key1",
				q: &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{
					{
						NsId: "ns1",
						Rows: []*protoqueryservice.Row{
							{
								Key:     []byte("key1"),
								Value:   []byte("hello"),
								Version: types.VersionNumber(0).Bytes(),
							},
						},
					},
				}},
				expected: &driver.VaultValue{Raw: []byte("hello"), Version: types.VersionNumber(0).Bytes()},
			},
			{
				ns:  "ns1",
				key: "key2",
				q: &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{
					{
						NsId: "ns1",
						Rows: []*protoqueryservice.Row{
							{
								Key:     []byte("key2"),
								Value:   []byte("hello"),
								Version: types.VersionNumber(0).Bytes(),
							},
						},
					},
				}},
				expected: &driver.VaultValue{Raw: []byte("hello"), Version: types.VersionNumber(0).Bytes()},
			},
			{
				ns:  "ns1",
				key: "key2",
				q: &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{
					{
						NsId: "ns1",
						Rows: []*protoqueryservice.Row{
							{
								Key:     []byte("key2"),
								Value:   []byte(""),
								Version: types.VersionNumber(1).Bytes(),
							},
						},
					},
				}},
				expected: &driver.VaultValue{Raw: []byte(""), Version: types.VersionNumber(1).Bytes()},
			},
		}

		for _, tc := range table {
			fake.GetRowsReturns(tc.q, nil)
			resp, err := qs.GetState(tc.ns, tc.key)
			require.NoError(t, err)
			require.Equal(t, tc.expected, resp)
		}
	})

	t.Run("GetState does not exist", func(t *testing.T) {
		t.Parallel()
		qs, fake := setupTest(t)

		table := []struct {
			ns       string
			key      string
			q        *protoqueryservice.Rows
			expected *driver.VaultValue
		}{
			{"ns1", "key1", nil, nil},
			{"ns1", "key1", &protoqueryservice.Rows{}, nil},
			{"ns1", "key1", &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{}}, nil},
			{"ns1", "key1", &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{{NsId: "ns1"}}}, nil},
			{"ns1", "key1", &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{{NsId: "ns1", Rows: []*protoqueryservice.Row{}}}}, nil},
		}

		for _, tc := range table {
			fake.GetRowsReturns(tc.q, nil)
			resp, err := qs.GetState(tc.ns, tc.key)
			require.NoError(t, err)
			require.Equal(t, tc.expected, resp)
		}
	})

	t.Run("GetState invalid query inputs", func(t *testing.T) {
		t.Parallel()
		qs, _ := setupTest(t)

		table := []struct {
			ns  string
			key string
		}{
			{"", ""},
			{"", "key1"},
			{"ns1", ""},
		}

		for _, tc := range table {
			_, err := qs.GetState(tc.ns, tc.key)
			require.Error(t, err)
		}
	})

	t.Run("GetStates happy path", func(t *testing.T) {
		t.Parallel()
		qs, fake := setupTest(t)

		table := []struct {
			m        map[driver.Namespace][]driver.PKey
			q        *protoqueryservice.Rows
			expected map[driver.Namespace]map[driver.PKey]driver.VaultValue
		}{
			{
				m: map[driver.Namespace][]driver.PKey{"ns1": {"key1"}},
				q: &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{
					{
						NsId: "ns1",
						Rows: []*protoqueryservice.Row{
							{
								Key:     []byte("key1"),
								Value:   []byte("hello"),
								Version: types.VersionNumber(0).Bytes(),
							},
						},
					},
				}},
				expected: map[driver.Namespace]map[driver.PKey]driver.VaultValue{
					"ns1": {
						"key1": driver.VaultValue{Raw: []byte("hello"), Version: types.VersionNumber(0).Bytes()},
					},
				},
			},
			{
				m: map[driver.Namespace][]driver.PKey{"ns1": {"key1", "key2"}},
				q: &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{
					{
						NsId: "ns1",
						Rows: []*protoqueryservice.Row{
							{
								Key:     []byte("key1"),
								Value:   []byte("hello"),
								Version: types.VersionNumber(0).Bytes(),
							},
							{
								Key:     []byte("key2"),
								Value:   []byte("hello2"),
								Version: types.VersionNumber(0).Bytes(),
							},
						},
					},
				}},
				expected: map[driver.Namespace]map[driver.PKey]driver.VaultValue{
					"ns1": {
						"key1": driver.VaultValue{Raw: []byte("hello"), Version: types.VersionNumber(0).Bytes()},
						"key2": driver.VaultValue{Raw: []byte("hello2"), Version: types.VersionNumber(0).Bytes()},
					},
				},
			},
			{
				m: map[driver.Namespace][]driver.PKey{"ns1": {"key1", "key2"}, "ns2": {"key3"}},
				q: &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{
					{
						NsId: "ns1",
						Rows: []*protoqueryservice.Row{
							{
								Key:     []byte("key1"),
								Value:   []byte("hello"),
								Version: types.VersionNumber(0).Bytes(),
							},
							{
								Key:     []byte("key2"),
								Value:   []byte("hello2"),
								Version: types.VersionNumber(0).Bytes(),
							},
						},
					},
					{
						NsId: "ns2",
						Rows: []*protoqueryservice.Row{
							{
								Key:     []byte("key3"),
								Value:   []byte("hello"),
								Version: types.VersionNumber(0).Bytes(),
							},
						},
					},
				}},
				expected: map[driver.Namespace]map[driver.PKey]driver.VaultValue{
					"ns1": {
						"key1": driver.VaultValue{Raw: []byte("hello"), Version: types.VersionNumber(0).Bytes()},
						"key2": driver.VaultValue{Raw: []byte("hello2"), Version: types.VersionNumber(0).Bytes()},
					},
					"ns2": {
						"key3": driver.VaultValue{Raw: []byte("hello"), Version: types.VersionNumber(0).Bytes()},
					},
				},
			},
		}

		for _, tc := range table {
			fake.GetRowsReturns(tc.q, nil)
			resp, err := qs.GetStates(tc.m)
			require.NoError(t, err)
			require.Equal(t, tc.expected, resp)
		}
	})

	t.Run("GetStates some do not exist", func(t *testing.T) {
		t.Parallel()
		qs, fake := setupTest(t)

		table := []struct {
			m        map[driver.Namespace][]driver.PKey
			q        *protoqueryservice.Rows
			expected map[driver.Namespace]map[driver.PKey]driver.VaultValue
		}{
			{
				m:        map[driver.Namespace][]driver.PKey{"ns1": {"key1"}},
				q:        &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{{NsId: "ns1"}}},
				expected: map[driver.Namespace]map[driver.PKey]driver.VaultValue{"ns1": {}},
			},
			{
				m: map[driver.Namespace][]driver.PKey{"ns1": {"key1", "doesnotexist"}},
				q: &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{
					{
						NsId: "ns1",
						Rows: []*protoqueryservice.Row{
							{
								Key:     []byte("key1"),
								Value:   []byte("hello"),
								Version: types.VersionNumber(0).Bytes(),
							},
						},
					},
				}},
				expected: map[driver.Namespace]map[driver.PKey]driver.VaultValue{
					"ns1": {
						"key1": driver.VaultValue{Raw: []byte("hello"), Version: types.VersionNumber(0).Bytes()},
					},
				},
			},
			{
				m: map[driver.Namespace][]driver.PKey{"ns1": {"doesnotexist", "key2"}},
				q: &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{
					{
						NsId: "ns1",
						Rows: []*protoqueryservice.Row{
							{
								Key:     []byte("key2"),
								Value:   []byte("hello"),
								Version: types.VersionNumber(0).Bytes(),
							},
						},
					},
				}},
				expected: map[driver.Namespace]map[driver.PKey]driver.VaultValue{
					"ns1": {
						"key2": driver.VaultValue{Raw: []byte("hello"), Version: types.VersionNumber(0).Bytes()},
					},
				},
			},
			{
				m: map[driver.Namespace][]driver.PKey{"ns1": {"key1"}, "nsdoesnotexist": {"key2"}},
				q: &protoqueryservice.Rows{Namespaces: []*protoqueryservice.RowsNamespace{
					{
						NsId: "ns1",
						Rows: []*protoqueryservice.Row{
							{
								Key:     []byte("key1"),
								Value:   []byte("hello"),
								Version: types.VersionNumber(0).Bytes(),
							},
						},
					},
				}},
				expected: map[driver.Namespace]map[driver.PKey]driver.VaultValue{
					"ns1": {
						"key1": driver.VaultValue{Raw: []byte("hello"), Version: types.VersionNumber(0).Bytes()},
					},
				},
			},
		}

		for _, tc := range table {
			fake.GetRowsReturns(tc.q, nil)
			resp, err := qs.GetStates(tc.m)
			require.NoError(t, err)
			require.Equal(t, tc.expected, resp)
		}
	})

	t.Run("GetStates invalid query inputs", func(t *testing.T) {
		t.Parallel()
		qs, _ := setupTest(t)

		table := []struct {
			m map[driver.Namespace][]driver.PKey
		}{
			{nil},
			{map[driver.Namespace][]driver.PKey{}},
			{map[driver.Namespace][]driver.PKey{"": {""}}},
			{map[driver.Namespace][]driver.PKey{"": {"key1"}}},
			{map[driver.Namespace][]driver.PKey{"ns1": {}}},
			{map[driver.Namespace][]driver.PKey{"ns1": {""}}},
			{map[driver.Namespace][]driver.PKey{"ns1": {"key1", ""}}},
			{map[driver.Namespace][]driver.PKey{"ns1": {"key1", "key2", ""}}},
			{map[driver.Namespace][]driver.PKey{"ns1": {"", "key2", ""}}},
		}

		for _, tc := range table {
			_, err := qs.GetStates(tc.m)
			require.Error(t, err)
		}
	})

	t.Run("GetState/s client return error", func(t *testing.T) {
		t.Parallel()
		qs, fake := setupTest(t)

		expectedError := fmt.Errorf("some error")

		fake.GetRowsReturns(nil, expectedError)
		_, err := qs.GetState("ns", "key1")
		require.ErrorIs(t, err, expectedError)
	})
}
