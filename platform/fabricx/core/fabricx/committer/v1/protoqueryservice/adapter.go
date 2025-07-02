/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoqueryservice

import (
	"context"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	api "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoqueryservice"
	protoblocktx_rc_0_1 "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v1/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/utils"
	"google.golang.org/grpc"
)

type queryServiceClientProvider struct {
	adapter protoblocktx_rc_0_1.MappingService
}

func (p *queryServiceClientProvider) GetClient(conn *grpc.ClientConn) api.QueryServiceClient {
	return NewServiceAdapter(NewQueryServiceClient(conn), p.adapter)
}

func NewQueryServiceClientProvider(adapter protoblocktx_rc_0_1.MappingService) *queryServiceClientProvider {
	return &queryServiceClientProvider{adapter: adapter}
}

func NewServiceAdapter(s QueryServiceClient, adapter protoblocktx_rc_0_1.MappingService) *serviceAdapter {
	return &serviceAdapter{s: s, adapter: adapter}
}

type serviceAdapter struct {
	s       QueryServiceClient
	adapter protoblocktx_rc_0_1.MappingService
}

func (a *serviceAdapter) GetRows(ctx context.Context, in api.Query, opts ...grpc.CallOption) (api.Rows, error) {
	queryNamespaces := make([]*QueryNamespace, len(in.GetNamespaces()))
	for i, namespace := range in.GetNamespaces() {
		nsID, err := a.adapter.IDByName(namespace.GetNsId())
		if err != nil {
			return nil, errors.Wrapf(err, "could not map ns [%s]", namespace.GetNsId())
		}
		queryNamespaces[i] = &QueryNamespace{
			NsId: nsID,
			Keys: namespace.GetKeys(),
		}
	}
	rows, err := a.s.GetRows(ctx, &Query{
		View:       mapView(in.GetView()),
		Namespaces: queryNamespaces,
	}, opts...)
	if err != nil {
		return nil, err
	}
	namespaces := make([]api.RowsNamespace, len(rows.GetNamespaces()))
	for i, namespace := range rows.GetNamespaces() {
		nsID, err := a.adapter.NameByID(namespace.GetNsId())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to map ns id [%d]", namespace.GetNsId())
		}
		namespaces[i] = api.NewRowsNamespace(nsID, utils.Map(namespace.GetRows(), func(r *Row) api.Row { return r }))
	}
	return api.NewRows(namespaces), nil
}

func (a *serviceAdapter) BeginView(ctx context.Context, in api.ViewParameters, opts ...grpc.CallOption) (api.View, error) {
	return a.s.BeginView(ctx, &ViewParameters{
		IsoLevel:            IsoLevel(in.GetIsoLevel()),
		NonDeferrable:       in.GetNonDeferrable(),
		TimeoutMilliseconds: uint64(in.GetTimeout().Milliseconds()), //nolint:gosec
	}, opts...)
}

func (a *serviceAdapter) EndView(ctx context.Context, in api.View, opts ...grpc.CallOption) (api.View, error) {
	return a.s.EndView(ctx, mapView(in), opts...)
}

func (*serviceAdapter) GetPolicies(context.Context, ...grpc.CallOption) (api.Policies, error) {
	panic("not implemented")
}

func mapView(view api.View) *View {
	if view == nil {
		return nil
	}
	return &View{Id: view.GetId()}
}
