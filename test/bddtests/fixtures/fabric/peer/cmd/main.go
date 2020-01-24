/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"strings"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/peer/node"
	viper "github.com/spf13/viper2015"
	extpeer "github.com/trustbloc/fabric-peer-ext/pkg/peer"
)

var logger = flogging.MustGetLogger("peer-ext-test")

func main() {
	setup()

	extpeer.Initialize()

	if err := startPeer(); err != nil {
		panic(err)
	}
}

func setup() {
	// For environment variables.
	viper.SetEnvPrefix(node.CmdRoot)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	node.InitCmd(nil, nil)
}

func startPeer() error {
	return node.Start()
}
