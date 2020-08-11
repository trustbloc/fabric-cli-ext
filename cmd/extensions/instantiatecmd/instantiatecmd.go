/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package instantiatecmd

import (
	"github.com/hyperledger/fabric-cli/cmd/commands/common"
	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trustbloc/fabric-cli-ext/cmd/basecmd"
	extcommon "github.com/trustbloc/fabric-cli-ext/cmd/extensions/common"
)

const (
	use      = "instantiatecc"
	desc     = "Instantiates chaincode"
	longDesc = `
The instantiatecc command allows a client to instantiate a chaincode using custom collection types, such as DCAS, off-ledger, and transient data.
`
	examples = `
- Instantiate a chaincode with DCAS and off-ledger collections:
    $ ./fabric-cli extensions instantiatecc mycc v1 --collections-config [{"name":"coll1","type":"COL_DCAS","policy":"OR('Org1MSP.member','Org2MSP.member')","maxPeerCount":2,"requiredPeerCount":1,"timeToLive":"10m"},{"name":"coll2","type":"COL_OFFLEDGER","policy":"OR('IMPLICIT-ORG.member')"}]
`
)

const (
	msgCCInstantiated = "Successfully instantiated chaincode"
)

// New returns the instantiatecc command
func New(settings *environment.Settings) *cobra.Command {
	return newCmd(settings, nil)
}

func newCmd(settings *environment.Settings, p basecmd.FactoryProvider) *cobra.Command {
	c := &command{
		Command: basecmd.New(settings, p),
	}

	cmd := &cobra.Command{
		Use:     use,
		Short:   desc,
		Long:    longDesc,
		Example: examples,
		Args:    c.ParseArgs(),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return c.Validate()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.run()
		},
	}

	c.Settings = settings
	cmd.SetOutput(c.Settings.Streams.Out)
	cmd.SilenceUsage = true

	c.AddArg(&c.ccName)
	c.AddArg(&c.ccVersion)

	flags := cmd.Flags()
	flags.StringVar(&c.ccPolicy, "policy", "", "sets the endorsement policy")
	flags.StringVar(&c.ccCollectionsConfig, "collections-config", "", "set the collections config (in JSON format)")

	cmd.SetOutput(c.Settings.Streams.Out)

	return cmd
}

// command implements the query command
type command struct {
	*basecmd.Command

	ccName              string
	ccVersion           string
	ccPolicy            string
	ccCollectionsConfig string
}

// Validate checks the required parameters for run
func (c *command) Validate() error {
	if len(c.ccName) == 0 {
		return errors.New("chaincode name not specified")
	}

	if len(c.ccVersion) == 0 {
		return errors.New("chaincode version not specified")
	}

	return nil
}

func (c *command) run() error {
	context, err := c.Settings.Config.GetCurrentContext()
	if err != nil {
		return err
	}

	policy, err := common.GetChaincodePolicy(c.ccPolicy)
	if err != nil {
		return err
	}

	collectionsConfig, err := extcommon.UnmarshalCollectionsConfig(c.ccCollectionsConfig)
	if err != nil {
		return err
	}

	req := resmgmt.InstantiateCCRequest{
		Name:       c.ccName,
		Version:    c.ccVersion,
		Policy:     policy,
		CollConfig: collectionsConfig,
		Path:       "not used",
	}

	options := []resmgmt.RequestOption{
		resmgmt.WithTargetEndpoints(context.Peers...),
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	}

	resMgmt, err := c.ResMgmt()
	if err != nil {
		return err
	}

	if _, err := resMgmt.InstantiateCC(context.Channel, req, options...); err != nil {
		return err
	}

	if err := c.Fprintln(c.Settings.Streams.Out, msgCCInstantiated); err != nil {
		return err
	}

	return nil
}
