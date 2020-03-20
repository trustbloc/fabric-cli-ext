/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package updatecmd

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/common"
)

type configPreProcessor struct {
	configFilePath string
}

func newConfigPreProcessor(configFilePath string) *configPreProcessor {
	return &configPreProcessor{configFilePath: configFilePath}
}

func (cp *configPreProcessor) preProcess(cfg *common.Config) (*common.Config, error) {
	peers, err := cp.visitPeers(cfg.Peers)
	if err != nil {
		return nil, err
	}
	apps, err := cp.visitApps(cfg.Apps)
	if err != nil {
		return nil, err
	}
	return &common.Config{
		MspID: cfg.MspID,
		Peers: peers,
		Apps:  apps,
	}, nil
}

func (cp *configPreProcessor) visitPeers(srcPeers []*common.Peer) ([]*common.Peer, error) {
	peers := make([]*common.Peer, len(srcPeers))
	for i, p := range srcPeers {
		apps, err := cp.visitApps(p.Apps)
		if err != nil {
			return nil, err
		}
		peers[i] = &common.Peer{
			PeerID: p.PeerID,
			Apps:   apps,
		}
	}
	return peers, nil
}

func (cp *configPreProcessor) visitApps(srcApps []*common.App) ([]*common.App, error) {
	apps := make([]*common.App, len(srcApps))
	for i, a := range srcApps {
		var config string
		var components []*common.Component
		var err error
		if a.Config != "" {
			config, err = cp.visitConfigString(a.Config)
			if err != nil {
				return nil, err
			}
		}

		components, err = cp.visitComponents(a.Components)
		if err != nil {
			return nil, err
		}

		apps[i] = &common.App{
			AppName:    a.AppName,
			Version:    a.Version,
			Format:     a.Format,
			Tags:       a.Tags,
			Config:     config,
			Components: components,
		}
	}
	return apps, nil
}

func (cp *configPreProcessor) visitComponents(srcComponents []*common.Component) ([]*common.Component, error) {
	components := make([]*common.Component, len(srcComponents))
	for i, c := range srcComponents {
		config, err := cp.visitConfigString(c.Config)
		if err != nil {
			return nil, err
		}
		components[i] = &common.Component{
			Name:    c.Name,
			Version: c.Version,
			Format:  c.Format,
			Tags:    c.Tags,
			Config:  config,
		}
	}
	return components, nil
}

func (cp *configPreProcessor) visitConfigString(srcConfig string) (string, error) {
	// Substitute all of the file refs with the actual contents of the file
	if !strings.HasPrefix(srcConfig, "file://") {
		return srcConfig, nil
	}

	refFilePath := srcConfig[7:]
	contents, err := cp.readFileRef(refFilePath)
	if err != nil {
		return "", errors.Wrapf(err, "error retrieving contents of file [%s]", refFilePath)
	}
	return string(contents), nil
}

func (cp *configPreProcessor) readFileRef(refPath string) ([]byte, error) {
	var path string
	if filepath.IsAbs(refPath) || cp.configFilePath == "" {
		path = refPath
	} else {
		// The path is relative to the source config file
		path = filepath.Join(filepath.Dir(cp.configFilePath), refPath)
	}
	return ioutil.ReadFile(filepath.Clean(path))
}
