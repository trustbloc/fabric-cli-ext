// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/fabric-cli-ext/test/bddtests

require (
	github.com/cucumber/godog v0.8.1
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta2.0.20200807145244-92e563b57775
	github.com/pkg/errors v0.8.1
	github.com/spf13/viper v1.4.0
	github.com/trustbloc/fabric-peer-test-common v0.1.4-0.20200811140027-026e52e34fa9
)

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.4-0.20200626180529-18936b36feca

go 1.13
