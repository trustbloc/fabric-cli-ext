/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package commitcmd

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
	use      = "commitcc"
	desc     = "Commits an approved chaincode"
	longDesc = `
The commitcc command allows a client to commit an approved chaincode using custom collection types, such as DCAS, off-ledger, and transient data.
`
	examples = `
- Commit a chaincode with DCAS and off-ledger collections:
    $ ./fabric-cli extensions commitcc mycc v1 1 --collections-config [{"name":"coll1","type":"COL_DCAS","policy":"OR('Org1MSP.member','Org2MSP.member')","maxPeerCount":2,"requiredPeerCount":1,"timeToLive":"10m"},{"name":"coll2","type":"COL_OFFLEDGER","policy":"OR('IMPLICIT-ORG.member')"}]
`
)

const (
	msgCCCommitted = "Successfully committed chaincode"
)

// New returns the commitcc command
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
	c.AddArg(&c.sequence)

	flags := cmd.Flags()
	flags.StringVar(&c.signaturePolicy, "policy", "", "sets the endorsement policy")
	flags.StringVar(&c.channelConfigPolicy, "channel-config-policy", "", "sets the channel config policy")
	flags.StringVar(&c.collectionsConfig, "collections-config", "", "set the collections config (in JSON format)")
	flags.BoolVar(&c.initRequired, "init-required", false, "indicates whether the chaincode requires 'Init' to be invoked")
	flags.StringVar(&c.endorsementPlugin, "endorsement-plugin", "", "sets the endorsement plugin")
	flags.StringVar(&c.validationPlugin, "validation-plugin", "", "sets the validation plugin")
	flags.StringArrayVar(&c.peers, "peer", []string{}, "sets a peer to which to send the commit (note that this option may be specified multiple times)")

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
	sequence            string
	initRequired        bool
	endorsementPlugin   string
	validationPlugin    string
	peers               []string
}

// Validate checks the required parameters for run
func (c *command) Validate() error {
	if len(c.name) == 0 {
		return errors.New("chaincode name not specified")
	}

	if len(c.version) == 0 {
		return errors.New("chaincode version not specified")
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

	peers := c.peers
	if len(peers) == 0 {
		peers = context.Peers
	}

	req := resmgmt.LifecycleCommitCCRequest{
		Name:                c.name,
		Version:             c.version,
		Sequence:            sequence,
		SignaturePolicy:     signaturePolicy,
		ChannelConfigPolicy: c.channelConfigPolicy,
		CollectionConfig:    collectionsConfig,
		InitRequired:        c.initRequired,
		EndorsementPlugin:   c.endorsementPlugin,
		ValidationPlugin:    c.validationPlugin,
	}

	options := []resmgmt.RequestOption{
		resmgmt.WithTargetEndpoints(peers...),
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	}

	resMgmt, err := c.ResMgmt()
	if err != nil {
		return err
	}

	if _, err := resMgmt.LifecycleCommitCC(context.Channel, req, options...); err != nil {
		return err
	}

	if err := c.Fprintln(c.Settings.Streams.Out, msgCCCommitted); err != nil {
		return err
	}

	return nil
}
