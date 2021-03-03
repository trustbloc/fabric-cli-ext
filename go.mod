// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/fabric-cli-ext

require (
	github.com/hyperledger/fabric-cli v0.0.0-20201005191300-d9e3966b20eb
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta3.0.20201002210629-a64e1ef9f926
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.6
	github.com/stretchr/testify v1.5.1
	github.com/trustbloc/sidetree-core-go v0.1.6-0.20210301232849-50c4792e1ca1
)

go 1.13

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.5

replace github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.3.0

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5

replace gopkg.in/check.v1 => gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127

replace github.com/stretchr/testify => github.com/stretchr/testify v1.5.1

replace github.com/spaolacci/murmur3 => github.com/spaolacci/murmur3 v1.1.0

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.3.0
