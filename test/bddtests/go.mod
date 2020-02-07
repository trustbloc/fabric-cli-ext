// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/fabric-cli-ext/test/bddtests

require (
	github.com/DATA-DOG/godog v0.7.13
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta1.0.20200203184105-5f7f0b025d89
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0 // indirect
	github.com/spf13/viper v1.4.0
	github.com/trustbloc/fabric-peer-test-common v0.1.1
)

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.1

go 1.13
