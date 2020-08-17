/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fileidxupdatecmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trustbloc/fabric-cli-ext/cmd/basecmd"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/common"
)

const (
	use      = "fileidxupdate"
	desc     = "Update the ID of the file index document for a given path"
	longDesc = `
The fileidxupdate command allows a client to update the file handler configuration of a peer with an ID of a Sidetree file index document`
	examples = `
- Updates the ID of the file index Sidetree document in two peers in Org1MSP:
    $ ./fabric ledgerconfig fileidxupdate --msp Org1MSP --peers peer0.org1.example.com;peer1.org1.example.com --path /content --idxid file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA== --noprompt
`
)

const (
	mspIDFlag  = "msp"
	mspIDUsage = `The ID of the MSP. Example: --msp Org1MSP`

	peersFlag  = "peers"
	peersUsage = "A semi-colon-separated list of peers. Example: --peers peer0.org1.com;peer1.org1.com"

	basePathFlag  = "path"
	basePathUsage = "The file handler path. Example: --path /schema"

	fileIndexIDFlag  = "idxid"
	fileIndexIDUsage = "The ID of the file index Sidetree document that is to be updated with the uploaded file. Example: --idxid file:idx:1234"

	noPromptFlag  = "noprompt"
	noPromptUsage = "If specified then operation will not prompt for confirmation. Example: --noprompt"

	msgConfigUpdated   = "File index successfully updated!"
	msgAborted         = "Operation aborted"
	msgContinueOrAbort = "Enter Y to continue or N to abort "
)

const (
	configSCC = "configscc"

	fileHandlerAppName          = "file-handler"
	fileHandlerAppVersion       = "1"
	fileHandlerComponentVersion = "1"
)

var (
	errMSPRequired         = errors.New("msp (--msp) is required")
	errPeersRequired       = errors.New("peers (--peers) is required")
	errFileIndexIDRequired = errors.New("file index ID (--idxid) is required")
	errBasePathRequired    = errors.New("base path (--path) is required")
)

// New returns the ledgerconfig update sub-command
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
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return c.run()
		},
	}

	c.Settings = settings
	cmd.SetOutput(c.Settings.Streams.Out)
	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&c.mspID, mspIDFlag, "", mspIDUsage)
	cmd.Flags().StringVar(&c.peerID, peersFlag, "", peersUsage)
	cmd.Flags().StringVar(&c.basePath, basePathFlag, "", basePathUsage)
	cmd.Flags().StringVar(&c.fileIndexID, fileIndexIDFlag, "", fileIndexIDUsage)
	cmd.Flags().BoolVar(&c.noPrompt, noPromptFlag, false, noPromptUsage)

	return cmd
}

// command implements the update command
type command struct {
	*basecmd.Command

	// Flags
	mspID       string
	peerID      string
	basePath    string
	fileIndexID string
	noPrompt    bool
}

func (c *command) validate() error {
	if c.mspID == "" {
		return errMSPRequired
	}

	if c.peerID == "" {
		return errPeersRequired
	}

	if c.basePath == "" {
		return errBasePathRequired
	}

	if c.fileIndexID == "" {
		return errFileIndexIDRequired
	}

	return nil
}

func (c *command) run() error {
	configBytes, err := c.getConfigBytes()
	if err != nil {
		return err
	}

	// Get confirmation from the user
	if !c.noPrompt {
		confirmed, e := c.confirmUpdate(configBytes)
		if e != nil {
			return e
		}
		if !confirmed {
			return c.Fprintln(msgAborted)
		}
	}

	req := channel.Request{
		ChaincodeID: common.ConfigSCC,
		Fcn:         "save",
		Args:        [][]byte{configBytes},
	}

	ch, err := c.Channel()
	if err != nil {
		return err
	}

	_, err = ch.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		return err
	}

	return c.Fprintln(msgConfigUpdated)
}

type authConfig struct {
	// ReadTokens contains a set of names of tokens for authorizing read requests
	ReadTokens []string
	// WriteTokens contains a set of names of tokens for authorizing write requests
	WriteTokens []string
}

type fileHandlerConfig struct {
	Authorization authConfig

	BasePath       string
	ChaincodeName  string
	Collection     string
	IndexNamespace string
	IndexDocID     string
}

func (c *command) getConfigBytes() ([]byte, error) {
	cfgMap, err := c.loadConfig()
	if err != nil {
		return nil, err
	}

	cfg := &common.Config{
		MspID: c.mspID,
	}

	for peerID, handlerCfg := range cfgMap {
		handlerCfg.IndexDocID = c.fileIndexID

		cfgBytes, err := json.Marshal(handlerCfg)
		if err != nil {
			return nil, err
		}

		peerCfg := &common.Peer{
			PeerID: peerID,
			Apps: []*common.App{
				{
					AppName: fileHandlerAppName,
					Version: fileHandlerAppVersion,
					Components: []*common.Component{
						{
							Name:    c.basePath,
							Version: fileHandlerComponentVersion,
							Format:  "JSON",
							Config:  string(cfgBytes),
						},
					},
				},
			},
		}

		cfg.Peers = append(cfg.Peers, peerCfg)
	}

	return json.Marshal(cfg)
}

func (c *command) loadConfig() (map[string]*fileHandlerConfig, error) {
	peers := strings.Split(c.peerID, ";")

	cfgMap := make(map[string]*fileHandlerConfig)
	for _, peerID := range peers {
		cfg, err := c.loadPeerConfig(peerID)
		if err != nil {
			return nil, err
		}

		if !strings.HasPrefix(c.fileIndexID, cfg.IndexNamespace) {
			return nil, errors.Errorf("file index ID must begin with [%s]", cfg.IndexNamespace)
		}

		if cfg.IndexDocID == c.fileIndexID {
			// Index already set for peerID. Skip this peerID
			continue
		}

		cfgMap[peerID] = cfg
	}

	if len(cfgMap) == 0 {
		return nil, errors.Errorf("the file index ID for [%s] is already set to [%s]", c.basePath, c.fileIndexID)
	}

	return cfgMap, nil
}

func (c *command) loadPeerConfig(peerID string) (*fileHandlerConfig, error) {
	criteria := &common.Criteria{
		MspID:            c.mspID,
		PeerID:           peerID,
		AppName:          fileHandlerAppName,
		AppVersion:       fileHandlerAppVersion,
		ComponentName:    c.basePath,
		ComponentVersion: fileHandlerComponentVersion,
	}

	criteriaBytes, err := json.Marshal(criteria)
	if err != nil {
		return nil, err
	}

	req := channel.Request{
		ChaincodeID: configSCC,
		Fcn:         "get",
		Args:        [][]byte{criteriaBytes},
	}

	ch, err := c.Channel()
	if err != nil {
		return nil, err
	}

	resp, err := ch.Query(req)
	if err != nil {
		return nil, err
	}

	var queryResults []*common.KeyValue
	err = json.Unmarshal(resp.Payload, &queryResults)
	if err != nil {
		return nil, err
	}

	if len(queryResults) == 0 {
		return nil, errors.Errorf("config not found for file handler [%s]", c.basePath)
	}

	cfg := &fileHandlerConfig{}

	err = json.Unmarshal([]byte(queryResults[0].Value.Config), cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// confirmUpdate prompts the user for confirmation of the update
func (c *command) confirmUpdate(config []byte) (bool, error) {
	displayedJSON, err := common.FormatJSON(config)
	if err != nil {
		return false, err
	}

	prompt := fmt.Sprintf("Updating the configuration with:\n\n%s\n\n%s", displayedJSON, msgContinueOrAbort)

	err = c.Fprintln(prompt)
	if err != nil {
		return false, err
	}

	return strings.ToLower(c.Prompt()) == "y", nil
}
