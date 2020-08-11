/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package approvecmd

import (
	"strconv"

	"github.com/hyperledger/fabric-cli/cmd/commands/common"
	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	extcommon "github.com/trustbloc/fabric-cli-ext/cmd/extensions/common"

	"github.com/trustbloc/fabric-cli-ext/cmd/basecmd"
)

const (
	use      = "approvecc"
	desc     = "Approves a chaincode"
	longDesc = `
The approvecc command allows a client to approve a chaincode using custom collection types, such as DCAS, off-ledger, and transient data.
`
	examples = `
- Approve a chaincode with DCAS and off-ledger collections:
    $ ./fabric-cli extensions approvecc mycc v1 mycc:12345 1 --collections-config [{"name":"coll1","type":"COL_DCAS","policy":"OR('Org1MSP.member','Org2MSP.member')","maxPeerCount":2,"requiredPeerCount":1,"timeToLive":"10m"},{"name":"coll2","type":"COL_OFFLEDGER","policy":"OR('IMPLICIT-ORG.member')"}]
`
)

const (
	msgCCApproved = "Successfully approved chaincode"
)

// New returns the approvecc command
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

	c.AddArg(&c.name)
	c.AddArg(&c.version)
	c.AddArg(&c.packageID)
	c.AddArg(&c.sequence)

	flags := cmd.Flags()
	flags.StringVar(&c.signaturePolicy, "policy", "", "sets the endorsement policy")
	flags.StringVar(&c.channelConfigPolicy, "channel-config-policy", "", "sets the channel config policy")
	flags.StringVar(&c.collectionsConfig, "collections-config", "", "set the collections config (in JSON format)")
	flags.BoolVar(&c.initRequired, "init-required", false, "indicates whether the chaincode requires 'Init' to be invoked")
	flags.StringVar(&c.endorsementPlugin, "endorsement-plugin", "", "sets the endorsement plugin")
	flags.StringVar(&c.validationPlugin, "validation-plugin", "", "sets the validation plugin")

	cmd.SetOutput(c.Settings.Streams.Out)

	return cmd
}

// command implements the approve command
type command struct {
	*basecmd.Command

	name                string
	version             string
	signaturePolicy     string
	channelConfigPolicy string
	collectionsConfig   string
	packageID           string
	sequence            string
	initRequired        bool
	endorsementPlugin   string
	validationPlugin    string
}

// Validate checks the required parameters for run
func (c *command) Validate() error {
	if c.name == "" {
		return errors.New("chaincode name not specified")
	}

	if c.version == "" {
		return errors.New("chaincode version not specified")
	}

	if c.packageID == "" {
		return errors.New("package ID not specified")
	}

	if c.sequence == "" {
		return errors.New("sequence not specified")
	}

	sequence, err := strconv.ParseInt(c.sequence, 10, 64)
	if err != nil {
		return errors.WithMessage(err, "invalid sequence")
	}

	if sequence <= 0 {
		return errors.New("sequence must be greater than 0")
	}

	return nil
}

func (c *command) run() error {
	context, err := c.Settings.Config.GetCurrentContext()
	if err != nil {
		return err
	}

	signaturePolicy, err := common.GetChaincodePolicy(c.signaturePolicy)
	if err != nil {
		return err
	}

	collectionsConfig, err := extcommon.UnmarshalCollectionsConfig(c.collectionsConfig)
	if err != nil {
		return err
	}

	sequence, err := strconv.ParseInt(c.sequence, 10, 64)
	if err != nil {
		return errors.WithMessage(err, "invalid sequence")
	}

	req := resmgmt.LifecycleApproveCCRequest{
		Name:                c.name,
		Version:             c.version,
		PackageID:           c.packageID,
		Sequence:            sequence,
		SignaturePolicy:     signaturePolicy,
		ChannelConfigPolicy: c.channelConfigPolicy,
		CollectionConfig:    collectionsConfig,
		InitRequired:        c.initRequired,
		EndorsementPlugin:   c.endorsementPlugin,
		ValidationPlugin:    c.validationPlugin,
	}

	options := []resmgmt.RequestOption{
		resmgmt.WithTargetEndpoints(context.Peers...),
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	}

	resMgmt, err := c.ResMgmt()
	if err != nil {
		return err
	}

	if _, err := resMgmt.LifecycleApproveCC(context.Channel, req, options...); err != nil {
		return err
	}

	if err := c.Fprintln(c.Settings.Streams.Out, msgCCApproved); err != nil {
		return err
	}

	return nil
}
