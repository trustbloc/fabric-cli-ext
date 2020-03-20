/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/spf13/cobra"

	"github.com/hyperledger/fabric-cli/pkg/environment"

	"github.com/trustbloc/fabric-cli-ext/cmd/file/createidxcmd"
)

const (
	use      = "file"
	desc     = "Manages file uploads"
	longDesc = "The file command allows you to upload files and create file indexes as Sidetree documents"
)

// New is the entry point to the file plugin
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
		createidxcmd.New(settings),
	)

	return cmd
}
