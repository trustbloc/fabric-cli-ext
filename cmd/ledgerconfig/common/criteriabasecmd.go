/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"encoding/json"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

// CriteriaBaseCommand may be used as a BaseCommand for commands that use search criteria
type CriteriaBaseCommand struct {
	*BaseCommand

	// Flags
	criteriaStr      string
	mspID            string
	peerID           string
	appName          string
	appVersion       string
	componentName    string
	componentVersion string
}

// NewCriteriaBaseCommand returns a CriteriaBaseCommand
func NewCriteriaBaseCommand(settings *environment.Settings, p FactoryProvider, cmd *cobra.Command) *CriteriaBaseCommand {
	c := &CriteriaBaseCommand{
		BaseCommand: NewBaseCmd(settings, p),
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

	return c
}

// Validate validates the flags
func (c *CriteriaBaseCommand) Validate() error {
	if c.criteriaStr != "" {
		if c.mspID != "" || c.peerID != "" || c.appName != "" || c.appVersion != "" || c.componentName != "" || c.componentVersion != "" {
			return errors.New(errCriteriaMustBeAlone)
		}

		// Validate the criteria
		criteria := &Criteria{}
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

// GetCriteriaBytes returns the Criteria marshalled as JSON
func (c *CriteriaBaseCommand) GetCriteriaBytes() ([]byte, error) {
	if c.criteriaStr != "" {
		return []byte(c.criteriaStr), nil
	}

	criteria := &Criteria{
		MspID:            c.mspID,
		PeerID:           c.peerID,
		AppName:          c.appName,
		AppVersion:       c.appVersion,
		ComponentName:    c.componentName,
		ComponentVersion: c.componentVersion,
	}
	return json.Marshal(criteria)
}

// GetConfig returns the config according to the given criteria
func (c *CriteriaBaseCommand) GetConfig(criteria []byte) ([]byte, error) {
	req := channel.Request{
		ChaincodeID: ConfigSCC,
		Fcn:         "get",
		Args:        [][]byte{criteria},
	}

	ch, err := c.Channel()
	if err != nil {
		return nil, err
	}

	resp, err := ch.Query(req)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
