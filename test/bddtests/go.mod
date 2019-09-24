// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/fabric-cli-ext/test/bddtests

require (
	github.com/DATA-DOG/godog v0.7.13
	github.com/hyperledger/fabric-sdk-go v1.0.0-alpha5.0.20190429134815-48bb0d199e2c
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0 // indirect
	github.com/spf13/viper v1.4.0
	github.com/trustbloc/fabric-peer-test-common v0.0.0-20190920153738-13a7665d089d
)

replace github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric => github.com/trustbloc/fabric-sdk-go-ext/fabric v0.0.0-20190528182243-b95c24511993
