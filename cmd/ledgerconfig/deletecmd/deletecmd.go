/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deletecmd

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/spf13/cobra"

	"github.com/trustbloc/fabric-cli-ext/cmd/basecmd"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/common"
)

const (
	use      = "delete"
	desc     = "Delete ledger configuration"
	longDesc = `
The delete command allows the client to delete the MSP's configuration using search criteria. The criteria consists of:

* MspID (mandatory)           - The MSP ID of the organization
* PeerID (optional)           - The ID of the peer
* AppName (optional)          - The application name
* AppVersion (optional)       - The application version
* ComponentName (optional)    - The component name
* ComponentVersion (optional) - The component version

Criteria may be specified as a JSON string (using the --criteria option) or it may be specified using the options:
	--mspid, --peerid, --appname, --appver, --componentname and --componentver

If PeerID and AppName are not specified then all of the MSP's configuration is deleted.
`
	examples = `
- Delete an application's configuration for a given peer:
    $ ./fabric ledgerconfig delete --mspid Org1MSP --peerid peer0.org1.com --appname myapp --appver 1

- Delete a specific version of a component's configuration':
    $ ./fabric ledgerconfig delete --mspid Org1MSP --appname myapp --appver 1 --componentname comp1 --componentver 1

- Delete all configuration in Org1MSP:
    $ ./fabric ledgerconfig delete --mspid Org1MSP
`
)

const (
	noPromptFlag        = "noprompt"
	noPromptDescription = "If specified then delete operation will not prompt for confirmation. Example: --noprompt"

	msgConfigDeleted   = "Configuration successfully deleted!"
	msgAborted         = "Operation aborted"
	msgContinueOrAbort = "Enter Y to continue or N to abort "
	msgNoConfig        = "No configuration matches the given criteria"
)

// New creates a new delete command
func New(settings *environment.Settings) *cobra.Command {
	return newCmd(settings, nil)
}

func newCmd(settings *environment.Settings, p basecmd.FactoryProvider) *cobra.Command {
	c := &command{}

	cmd := &cobra.Command{
		Use:     use,
		Short:   desc,
		Long:    longDesc,
		Example: examples,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return c.Validate()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.run()
		},
	}
	c.CriteriaBaseCommand = common.NewCriteriaBaseCommand(settings, p, cmd)

	cmd.Flags().BoolVar(&c.noPrompt, noPromptFlag, false, noPromptDescription)

	return cmd
}

// command implements the delete command
type command struct {
	*common.CriteriaBaseCommand

	// Flags
	noPrompt bool
}

func (c *command) run() error {
	criteriaBytes, err := c.GetCriteriaBytes()
	if err != nil {
		return err
	}

	req := channel.Request{
		ChaincodeID: common.ConfigSCC,
		Fcn:         "delete",
		Args:        [][]byte{criteriaBytes},
	}

	// Get confirmation from the user
	if !c.noPrompt {
		// Display to the user the configuration that will be deleted
		config, e := c.GetConfig(criteriaBytes)
		if e != nil {
			return e
		}
		if string(config) == "null" {
			return c.Fprintln(msgNoConfig)
		}
		confirmed, e := c.confirmDelete(config)
		if e != nil {
			return e
		}
		if !confirmed {
			return c.Fprintln(msgAborted)
		}
	}

	ch, err := c.Channel()
	if err != nil {
		return err
	}

	_, err = ch.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		return err
	}

	return c.Fprintln(msgConfigDeleted)
}

// confirmDelete prompts the user for confirmation of the delete
func (c *command) confirmDelete(config []byte) (bool, error) {
	displayedJSON, err := common.FormatJSON(config)
	if err != nil {
		return false, err
	}
	prompt := fmt.Sprintf("The following configuration will be deleted:\n\n%s\n\n%s", displayedJSON, msgContinueOrAbort)
	err = c.Fprintln(prompt)
	if err != nil {
		return false, err
	}
	return strings.ToLower(c.Prompt()) == "y", nil
}
