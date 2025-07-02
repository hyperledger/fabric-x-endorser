/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package namespace

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/types"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/v2/protoqueryservice"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/vault/queryservice"
)

func List(config *queryservice.Config) error {
	conn, err := queryservice.GrpcClient(config)
	if err != nil {
		return fmt.Errorf("cannot get grpc client: %w", err)
	}
	defer conn.Close()

	client := protoqueryservice.NewServiceAdapter(protoqueryservice.NewQueryServiceClient(conn))

	ctx, c := context.WithTimeout(context.Background(), config.QueryTimeout)
	defer c()
	res, err := client.GetPolicies(ctx)
	if err != nil {
		return fmt.Errorf("cannot query existing namespaces: %w", err)
	}

	fmt.Printf("Installed namespaces:\n")
	for i, p := range res.GetPolicies() {
		fmt.Printf("%d) %v: version %d policy: %x \n", i, p.GetNamespace(), types.VersionNumberFromBytes(p.GetVersion()), p.GetPolicy())
	}
	fmt.Printf("\n")

	return nil
}
