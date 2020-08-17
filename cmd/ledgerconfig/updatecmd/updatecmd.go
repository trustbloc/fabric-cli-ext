/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package updatecmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	use      = "update"
	desc     = "Update ledger configuration"
	longDesc = `
The update command allows a client to add/update the configuration of one or more applications.
Configuration can be specified directly on the command-line as a JSON string using the --config option,
or the path of a configuration file may be specified using the --configfile option. The configuration
string may be embedded directly in the "Config" element or the Config element may reference a file
containing the configuration.

The format of the configuration for config with peer is as follows:

{
  "MspID": "msp.one",
  "Peers": [
    {
      "PeerID": "peer1",
      "App": [
        {
          "AppName": "app1",
          "Version": "1",
          "Format":"JSON",
          "Config": "{\"Key1\":\"value1\",\"Key2\":\"Value2\"}"
        },
        {
          "AppName": "app1",
          "Version": "2",
          "Format":"JSON",
          "Config": "{\"Key1\":\"value1_1\",\"Key2\":\"Value2_1\"}"
        },
        {
          "AppName": "app2",
          "Version": "1",
          "Format":"YAML",
          "Config": "file://path_to_config.yaml"
        }
      ]
    },
    {
      "PeerID":"peer2",
      . . .
	}
  ]
}

The format of the configuration for peer-less config is displayed below:
{
  "MspID": "Org1MSP",
  "Apps": [
    {
      "AppName": "app1",
      "Version": "1",
      "Format":"Other",
      "Config": "{config goes here}"
    },
    {
      "AppName": "app2",
      "Version": "1",
      "Components": [
        {
          "Name": "comp1",
          "Version": "1"
          "Format":"Other",
          "Config": "{comp1 data ver 1}",
        },
        {
          "Name": "comp1",
          "Version": "2"
          "Format":"Other",
          "Config": "{comp1 data ver 2}",
        },
        {
          "Name": "comp2",
          "Version": "1"
          "Format":"Other",
          "Config": "{comp2 data ver 1}",
        }
      ]
    },
    .....
  ]
}
`
	examples = `
- Send the update using a configuration file:
    $ ./fabric ledgerconfig update --configfile ./sampleconfig/org1-config.json

- Send an update using a configuration string specified in the command-line:
    $ ./fabric ledgerconfig update --config '{"MspID":"Org1MSP","Peers":[{"PeerID":"peer0.org1.com","App":[{"AppName":"app1","Version":"v1","Format":"Other","Config":"embedded config"}]}]}'

- Send an update using a peer-less configuration string specified in the command-line:
    $ ./fabric ledgerconfig update --config '{"MspID":"Org1MSP","Apps":[{"AppName":"app1","Version":"v1","Format":"Other","Config":"embedded config"}]}'

- Send an update using a peer-less configuration string specified in the command-line:
    $ ./fabric ledgerconfig update --config '{"MspID":"general", "Apps": [{"AppName": "publickey", "Version": "v1", "Components": [{"Name":"comp1","Format":"Other","Config":"config1"}] }]}'
`
)

const (
	configFlag  = "config"
	configUsage = `The config update string in JSON format. Example: --config '{"MspID":"Org1MSP","Peers":[{"PeerID":"peer0.org1.com","App":[{"AppName":"myapp","Version":"1","Format":"JSON",Config":"{\"Org\":\"Org1MSP\",\"Application\":\"app1\"}"}]}]}'`

	configFileFlag  = "configfile"
	configFileUsage = `The path to the config file. Example: --configfile "./configs/msp1_config.json"`

	noPromptFlag  = "noprompt"
	noPromptUsage = "If specified then update operation will not prompt for confirmation. Example: --noprompt"

	errConfigOrConfigFileRequired = "one of --config or --configfile must be specified"
	errInvalidJSONConfig          = "invalid JSON config"
	errFileNotFound               = "file not found"

	msgConfigUpdated   = "Configuration successfully updated!"
	msgAborted         = "Operation aborted"
	msgContinueOrAbort = "Enter Y to continue or N to abort "
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

	cmd.Flags().StringVar(&c.config, configFlag, "", configUsage)
	cmd.Flags().StringVar(&c.configFile, configFileFlag, "", configFileUsage)
	cmd.Flags().BoolVar(&c.noPrompt, noPromptFlag, false, noPromptUsage)

	return cmd
}

// command implements the update command
type command struct {
	*basecmd.Command

	// Flags
	config     string
	configFile string
	noPrompt   bool
}

func (c *command) validate() error {
	if (c.config == "" && c.configFile == "") || (c.config != "" && c.configFile != "") {
		return errors.New(errConfigOrConfigFileRequired)
	}
	if c.config != "" {
		return validateConfig(c.config)
	}
	return validateConfigFile(c.configFile)
}

func (c *command) run() error {
	configBytes, err := c.getConfigBytes()
	if err != nil {
		return err
	}

	cfg := &common.Config{}
	err = json.Unmarshal(configBytes, cfg)
	if err != nil {
		return err
	}

	// Replace all of the file references with actual config
	newCfg, err := newConfigPreProcessor(c.configFile).preProcess(cfg)
	if err != nil {
		return err
	}

	configBytes, err = json.Marshal(newCfg)
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

func (c *command) getConfigBytes() ([]byte, error) {
	if c.config != "" {
		return []byte(c.config), nil
	}
	return ioutil.ReadFile(filepath.Clean(c.configFile))
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

func validateConfig(cfg string) error {
	config := &common.Config{}
	if err := json.Unmarshal([]byte(cfg), config); err != nil {
		return errors.WithMessagef(err, errInvalidJSONConfig)
	}
	return nil
}

func validateConfigFile(file string) error {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return errors.Errorf("%s: [%s]", errFileNotFound, file)
	}
	return nil
}
