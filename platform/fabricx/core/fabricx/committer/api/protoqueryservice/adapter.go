/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoqueryservice

import (
	"context"

	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/types"
	"google.golang.org/grpc"
)

type Provider = types.Provider[QueryServiceClientProvider]

func NewProvider(configService driver.ConfigService) (*types.ServiceProvider[QueryServiceClientProvider], error) {
	return types.NewProvider[QueryServiceClientProvider](configService)
}

type QueryServiceClientProvider interface {
	GetClient(*grpc.ClientConn) QueryServiceClient
}

// QueryServiceClient is the client API for QueryService service.
type QueryServiceClient interface {
	GetRows(ctx context.Context, in Query, opts ...grpc.CallOption) (Rows, error)
	BeginView(ctx context.Context, in ViewParameters, opts ...grpc.CallOption) (View, error)
	EndView(ctx context.Context, in View, opts ...grpc.CallOption) (View, error)
	GetPolicies(ctx context.Context, opts ...grpc.CallOption) (Policies, error)
}
