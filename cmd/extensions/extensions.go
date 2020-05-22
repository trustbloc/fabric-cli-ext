/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/spf13/cobra"

	"github.com/trustbloc/fabric-cli-ext/cmd/extensions/instantiatecmd"
)

const (
	use      = "extensions"
	desc     = "Fabric extensions"
	longDesc = "Provides extended functionality to work with fabric-peer-ext."
)

// New is the entry point to the ext plugin
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
		instantiatecmd.New(settings),
	)
	return cmd
}
