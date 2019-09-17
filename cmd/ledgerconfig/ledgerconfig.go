/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/spf13/cobra"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/querycmd"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/updatecmd"
)

const (
	use      = "ledgerconfig"
	desc     = "Manages ledger configuration"
	longDesc = "The ledgerconfig command allows you to update, delete, and query ledger configuration."
)

// New is the entry point to the ledgerconfig plugin
func New(settings *environment.Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: desc,
		Long:  longDesc,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	cmd.AddCommand(
		querycmd.New(settings),
		updatecmd.New(settings),
	)
	return cmd
}
