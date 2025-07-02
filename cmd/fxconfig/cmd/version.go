/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"

	"github.com/hyperledger/fabric-x-endorser/cmd/fxconfig/internal"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of fxconfig",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("fxconfig")
			showLine(cmd, "Version", internal.Version)
			showLine(cmd, "Go version", internal.GoVersion)
			showLine(cmd, "Commit", internal.Commit)
			showLine(cmd, "OS/Arch", fmt.Sprintf("%s/%s", internal.Os, internal.Arch))
		},
	}

	return cmd
}

func showLine(cmd *cobra.Command, title, value string) {
	cmd.Printf(" %-16s %s\n", fmt.Sprintf("%s:", cases.Title(language.Und, cases.NoLower).String(title)), value)
}
