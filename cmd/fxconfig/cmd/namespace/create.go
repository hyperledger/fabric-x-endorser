/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package namespace

import (
	"errors"
	"time"

	"github.com/hyperledger/fabric-x-endorser/cmd/fxconfig/internal/namespace"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/types"
	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	var committerVersion string
	var ordererCfg namespace.OrdererConfig
	var mspCfg namespace.MSPConfig
	var pkPath string

	cmd := &cobra.Command{
		Use:   "create NAMESPACE_NAME",
		Short: "Create Namespace",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nsID := args[0]

			channelName, err := cmd.Flags().GetString("channel")
			if err != nil {
				return err
			}

			if channelName == "" {
				return errors.New("you must specify a channel name '--channel channelName'")
			}

			return namespace.DeployNamespace(channelName, nsID, 0, types.CommitterVersion(committerVersion), ordererCfg, mspCfg, pkPath)
		},
	}

	cmd.PersistentFlags().String("channel", "", "The name of the channel")
	cmd.PersistentFlags().StringVarP(&committerVersion, "committer-version", "", "", "The version of scalable committer to use")

	// adds flags for orderer-related commands
	cmd.PersistentFlags().StringVarP(&ordererCfg.OrderingEndpoint, "orderer", "o", "",
		"Ordering service endpoint")
	cmd.PersistentFlags().BoolVarP(&ordererCfg.TLSEnabled, "tls", "", false,
		"Use TLS when communicating with the orderer endpoint")
	cmd.PersistentFlags().BoolVarP(&ordererCfg.ClientAuth, "clientauth", "", false,
		"Use mutual TLS when communicating with the orderer endpoint")
	cmd.PersistentFlags().StringVarP(&ordererCfg.CaFile, "cafile", "", "",
		"Path to file containing PEM-encoded trusted certificate(s) for the ordering endpoint")
	cmd.PersistentFlags().StringVarP(&ordererCfg.KeyFile, "keyfile", "", "",
		"Path to file containing PEM-encoded private key to use for mutual TLS communication with the orderer endpoint")
	cmd.PersistentFlags().StringVarP(&ordererCfg.CertFile, "certfile", "", "",
		"Path to file containing PEM-encoded X509 public key to use for mutual TLS communication with the orderer endpoint")
	cmd.PersistentFlags().StringVarP(&ordererCfg.OrdererTLSHostnameOverride, "ordererTLSHostnameOverride", "", "",
		"The hostname override to use when validating the TLS connection to the orderer")
	cmd.PersistentFlags().DurationVarP(&ordererCfg.ConnTimeout, "connTimeout", "", 3*time.Second,
		"Timeout for client to connect")
	cmd.PersistentFlags().DurationVarP(&ordererCfg.TLSHandshakeTimeShift, "tlsHandshakeTimeShift", "", 0,
		"The amount of time to shift backwards for certificate expiration checks during TLS handshakes with the orderer endpoint")

	// adds flags to specify the MSP that will sign the requests
	cmd.PersistentFlags().StringVarP(&mspCfg.MSPConfigPath, "mspConfigPath", "", "", "The path to the MSP config directory")
	cmd.PersistentFlags().StringVarP(&mspCfg.MSPID, "mspID", "", "", "The name of the MSP")

	cmd.PersistentFlags().StringVarP(&pkPath, "pk", "", "", "The path to the public key of the endorser")

	return cmd
}
