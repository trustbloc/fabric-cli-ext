/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package querycmd

import (
	"encoding/json"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
- Query configuration of a particular application on in a specified peer:

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
	criteriaFlag  = "criteria"
	criteriaUsage = `The search criteria in JSON format. Example: --criteria '{"MspID":"Org1MSP","PeerID":"peer0.org1.com","AppName":"app1","AppVersion":"v1","ComponentName":"comp1","ComponentVersion":"v1"}'`

	mspIDFlag  = "mspid"
	mspIDUsage = `The ID of the MSP. Example: --mspid Org1MSP`

	peerIDFlag  = "peerid"
	peerIDUsage = "The ID of the peer to query for. Example: --peerid peer0.org1.com"

	appNameFlag  = "appname"
	appNameUsage = "The name of the application to query for. Example: --appname app1"

	appVerFlag  = "appver"
	appVerUsage = "The app version. Example: --appver v1"

	componentNameFlag  = "componentname"
	componentNameUsage = "The name of the component to query for. Example: --componentname comp1"

	componentVerFlag  = "componentver"
	componentVerUsage = "The component version. Example: --componentver v1"
)

var (
	errMspOrCriteriaRequired = "either --criteria or (at least) --mspid must be specified"
	errCriteriaMustBeAlone   = "other options cannot be used along with --criteria"
	errInvalidCriteria       = "invalid criteria"
)

// New returns the ledger config query command
func New(settings *environment.Settings) *cobra.Command {
	return newCmd(settings, nil)
}

func newCmd(settings *environment.Settings, p common.FactoryProvider) *cobra.Command {
	c := &command{
		BaseCommand: common.NewBaseCmd(settings, p),
	}
	c.Settings = settings

	cmd := &cobra.Command{
		Use:     use,
		Short:   desc,
		Long:    longDesc,
		Example: examples,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return c.validate()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.run()
		},
	}

	c.Settings = settings
	cmd.SetOutput(c.Settings.Streams.Out)
	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&c.criteriaStr, criteriaFlag, "", criteriaUsage)
	cmd.Flags().StringVar(&c.mspID, mspIDFlag, "", mspIDUsage)
	cmd.Flags().StringVar(&c.peerID, peerIDFlag, "", peerIDUsage)
	cmd.Flags().StringVar(&c.appName, appNameFlag, "", appNameUsage)
	cmd.Flags().StringVar(&c.appVersion, appVerFlag, "", appVerUsage)
	cmd.Flags().StringVar(&c.componentName, componentNameFlag, "", componentNameUsage)
	cmd.Flags().StringVar(&c.componentVersion, componentVerFlag, "", componentVerUsage)

	return cmd
}

// command implements the query command
type command struct {
	*common.BaseCommand

	// Flags
	criteriaStr      string
	mspID            string
	peerID           string
	appName          string
	appVersion       string
	componentName    string
	componentVersion string
}

func (c *command) validate() error {
	if c.criteriaStr != "" {
		if c.mspID != "" || c.peerID != "" || c.appName != "" || c.appVersion != "" || c.componentName != "" || c.componentVersion != "" {
			return errors.New(errCriteriaMustBeAlone)
		}

		// Validate the criteria
		criteria := &common.Criteria{}
		if err := json.Unmarshal([]byte(c.criteriaStr), criteria); err != nil {
			return errors.WithMessagef(err, errInvalidCriteria)
		}
	} else {
		if c.mspID == "" {
			return errors.New(errMspOrCriteriaRequired)
		}
	}
	return nil
}

func (c *command) run() error {
	var criteriaBytes []byte
	if c.criteriaStr != "" {
		criteriaBytes = []byte(c.criteriaStr)
	} else {
		criteria := &common.Criteria{
			MspID:            c.mspID,
			PeerID:           c.peerID,
			AppName:          c.appName,
			AppVersion:       c.appVersion,
			ComponentName:    c.componentName,
			ComponentVersion: c.componentVersion,
		}
		var err error
		criteriaBytes, err = json.Marshal(criteria)
		if err != nil {
			return err
		}
	}

	req := channel.Request{
		ChaincodeID: common.ConfigSCC,
		Fcn:         "get",
		Args:        [][]byte{criteriaBytes},
	}

	ch, err := c.Channel()
	if err != nil {
		return err
	}

	resp, err := ch.Query(req, channel.WithTargetEndpoints(c.Context().Peers...))
	if err != nil {
		return err
	}

	return c.Fprintln(string(resp.Payload))
}
