/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package querycmd

import (
	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/spf13/cobra"
	"github.com/trustbloc/fabric-cli-ext/cmd/basecmd"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/common"
)

const (
	use      = "query"
	desc     = "Query ledger configuration"
	longDesc = `
The query command allows the client to query an MSP's configuration using search criteria. The criteria consists of:

* MspID (mandatory)           - The MSP ID of the organization
* PeerID (optional)           - The ID of the peer
* AppName (optional)          - The application name
* AppVersion (optional)       - The application version
* ComponentName (optional)    - The component name
* ComponentVersion (optional) - The component version

Criteria may be specified as a JSON string (using the --criteria option) or it may be specified using the options:
	--mspid, --peerid, --appname, --appver, --componentname and --componentver

If PeerID and AppName are not specified then all of the MSP's configuration is returned.
`
	examples = `
- Query configuration of a particular application on a specified peer:

    $ ./fabric ledgerconfig query --mspid Org1MSP --peerid peer0.org1.com --appname app1 --appver v1

... results in the following output:

	[{"MspID":"Org1MSP","PeerID":"peer0.org1.com","AppName":"myapp","AppVersion":"v1","TxID":"9730813e01479c0db5da31676dfe301eef1d7045ad3b5ea405d679c03c701433","Format":"JSON","Config":"{\"Org\":\"Org1MSP\",\"Application\":\"app1\"}"}]

- Query configuration of a particular application component:

    $ ./fabric ledgerconfig query --mspid Org1MSP --appname app2 --appver v1 --componentname comp1 --componentver v1

... results in the following output:

	[{"MspID":"Org1MSP","PeerID":"","AppName":"app2","AppVersion":"v1","ComponentName":"comp1","ComponentVersion":"v1","TxID":"9730813e01479c0db5da31676dfe301eef1d7045ad3b5ea405d679c03c701433","Format":"Other","Config":"some config"}]

- Query for all configuration in Org1MSP:

    $ ./fabric ledgerconfig query --mspid Org1MSP

- Query for configuration using JSON criteria:
    $ ./fabric ledgerconfig query --criteria '{"MspID":"Org1MSP","PeerID":"peer0.org1.com","AppName":"app1","AppVersion":"v1"}'
`
)

const (
	formatFlag  = "format"
	formatUsage = "If specified then displayed JSON will be formatted. Example: --format"

	msgNoConfig = "No configuration matches the given criteria"
)

// New returns the ledger config query command
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
	cmd.Flags().BoolVar(&c.formatJSON, formatFlag, false, formatUsage)
	c.CriteriaBaseCommand = common.NewCriteriaBaseCommand(settings, p, cmd)
	return cmd
}

// command implements the query command
type command struct {
	*common.CriteriaBaseCommand

	// Flags
	formatJSON bool
}

func (c *command) run() error {
	criteriaBytes, err := c.GetCriteriaBytes()
	if err != nil {
		return err
	}

	config, err := c.GetConfig(criteriaBytes)
	if err != nil {
		return err
	}

	var displayedJSON []byte
	if c.formatJSON {
		if string(config) == "null" {
			return c.Fprintln(msgNoConfig)
		}
		displayedJSON, err = common.FormatJSON(config)
		if err != nil {
			return err
		}
	} else {
		displayedJSON = config
	}

	return c.Fprintln(string(displayedJSON))
}
