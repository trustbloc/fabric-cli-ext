/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/cucumber/godog"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"
)

var logger = logging.NewLogger("test-logger")

const (
	networkName   = "test-network"
	sdkConfigPath = "./fixtures/config/sdk-client/config.yaml"
)

// FabricCLISteps extend the BDD test with Fabric CLI steps
type FabricCLISteps struct {
	BDDContext *bddtests.BDDContext
}

// NewFabricCLISteps returns fabric-cli BDD steps
func NewFabricCLISteps(context *bddtests.BDDContext) *FabricCLISteps {
	return &FabricCLISteps{BDDContext: context}
}

func (d *FabricCLISteps) installPlugin(path string) error {
	_, err := NewFabricCLI().Exec("plugin", "install", path)
	return err
}

func (d *FabricCLISteps) initNetwork() error {
	err := os.RemoveAll(homeDir)
	if err != nil {
		return err
	}
	out, err := NewFabricCLI().Exec("network", "set", networkName, sdkConfigPath)
	if err != nil {
		logger.Errorf("Error: %s:%s", err, out)
	}
	return err
}

func (d *FabricCLISteps) defineContext(name, channelID, orgID, strPeers, userID string) error {
	peers := strings.Split(strPeers, ",")
	if len(peers) == 0 {
		return errors.New("at least one peer must be specified")
	}

	var args []string
	args = append(args, "context", "set", name, "--network", networkName, "--channel", channelID, "--user", userID, "--organization", orgID)
	for _, peer := range peers {
		args = append(args, "--peers", peer)
	}

	_, err := NewFabricCLI().Exec(args...)
	return err
}

func (d *FabricCLISteps) useContext(name string) error {
	_, err := NewFabricCLI().Exec("context", "use", name)
	return err
}

func (d *FabricCLISteps) execute(strArgs string) error {
	bddtests.ClearResponse()

	args, err := bddtests.ResolveAllVars(strings.Replace(strArgs, " ", ",", -1))
	if err != nil {
		return err
	}
	logger.Infof("Executing fabric-cli with args: %s ...", args)
	response, err := NewFabricCLI().Exec(args...)
	if err != nil {
		return err
	}

	logger.Infof("... got response: %s", response)
	bddtests.SetResponse(response)
	return nil
}

func (d *FabricCLISteps) httpGet(url string) error {
	bddtests.ClearResponse()

	resolved, err := bddtests.ResolveVars(url)
	if err != nil {
		return err
	}

	url = resolved.(string)

	client := &http.Client{}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received status code %d", resp.StatusCode)
	}

	contentType, ok := resp.Header["Content-Type"]
	if ok && strings.HasPrefix(contentType[0], "image") {
		logger.Infof("... got HTTP image of type [%s]", contentType)
		return nil
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body failed: %s", err)
	}

	logger.Infof("... got HTTP response of type [%s]:\n%s", contentType[0], payload)

	bddtests.SetResponse(string(payload))

	return nil
}

// RegisterSteps registers transient data steps
func (d *FabricCLISteps) RegisterSteps(s *godog.Suite) {
	s.BeforeScenario(d.BDDContext.BeforeScenario)
	s.AfterScenario(d.BDDContext.AfterScenario)
	s.Step(`^fabric-cli network is initialized$`, d.initNetwork)
	s.Step(`^fabric-cli plugin "([^"]*)" is installed$`, d.installPlugin)
	s.Step(`^fabric-cli context "([^"]*)" is defined on channel "([^"]*)" with org "([^"]*)", peers "([^"]*)" and user "([^"]*)"$`, d.defineContext)
	s.Step(`^fabric-cli context "([^"]*)" is used$`, d.useContext)
	s.Step(`^fabric-cli is executed with args "([^"]*)"$`, d.execute)
	s.Step(`^an HTTP request is sent to "([^"]*)"$`, d.httpGet)
}
