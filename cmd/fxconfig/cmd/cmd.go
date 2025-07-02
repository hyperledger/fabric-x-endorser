/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"os"

	"github.com/hyperledger/fabric-x-endorser/cmd/fxconfig/cmd/namespace"
	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{Use: "fxconfig"}
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(namespace.NewNamespaceCommand())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
