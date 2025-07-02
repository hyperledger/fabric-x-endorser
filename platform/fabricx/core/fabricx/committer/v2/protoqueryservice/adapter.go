/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoqueryservice

import (
	"context"

	api "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoqueryservice"
	protoblocktx "github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v2/protoblocktx"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/utils"
	"google.golang.org/grpc"
)

type queryServiceClientProvider struct{}

func (p *queryServiceClientProvider) GetClient(conn *grpc.ClientConn) api.QueryServiceClient {
	return NewServiceAdapter(NewQueryServiceClient(conn))
}

func NewQueryServiceClientProvider() *queryServiceClientProvider {
	return &queryServiceClientProvider{}
}

func NewServiceAdapter(s QueryServiceClient) *serviceAdapter {
	return &serviceAdapter{s: s}
}

type serviceAdapter struct {
	s QueryServiceClient
}

func (a *serviceAdapter) GetRows(ctx context.Context, in api.Query, opts ...grpc.CallOption) (api.Rows, error) {
	rows, err := a.s.GetRows(ctx, &Query{
		View:       mapView(in.GetView()),
		Namespaces: utils.Map(in.GetNamespaces(), mapQueryNamespace),
	}, opts...)
	if err != nil {
		return nil, err
	}
	namespaces := utils.Map(rows.GetNamespaces(), mapRowsNamespace)
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

func mapView(view api.View) *View {
	if view == nil {
		return nil
	}
	return &View{Id: view.GetId()}
}

func (a *serviceAdapter) GetPolicies(ctx context.Context, opts ...grpc.CallOption) (api.Policies, error) {
	policies, err := a.s.GetPolicies(ctx, &Empty{}, opts...)
	if err != nil {
		return nil, err
	}
	return api.NewPolicies(utils.Map(policies.GetPolicies(), func(p *protoblocktx.PolicyItem) api.PolicyItem { return p })), nil
}

func mapRowsNamespace(namespace *RowsNamespace) api.RowsNamespace {
	return api.NewRowsNamespace(
		namespace.GetNsId(),
		utils.Map(namespace.GetRows(), func(r *Row) api.Row { return r }))
}

func mapQueryNamespace(n api.QueryNamespace) *QueryNamespace {
	return &QueryNamespace{NsId: n.GetNsId(), Keys: n.GetKeys()}
}
