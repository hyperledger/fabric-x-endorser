/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package namespace

import (
	"time"

	"github.com/hyperledger/fabric-x-endorser/cmd/fxconfig/internal/namespace"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/vault/queryservice"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	// TODO: check what parameters we need here
	config := &queryservice.Config{
		Endpoints: []queryservice.Endpoint{
			{
				Address:           "localhost:7001",
				ConnectionTimeout: 5 * time.Second,
			},
		},
		QueryTimeout: 10 * time.Second,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed Namespaces",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return namespace.List(config)
		},
	}

	cmd.PersistentFlags().StringVar(&config.Endpoints[0].Address, "endpoint", "", "committer query service endpoint")

	return cmd
}
