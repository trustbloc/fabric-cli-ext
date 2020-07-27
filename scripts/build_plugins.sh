#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

echo "Building fabric-cli..."

declare envOS
envOS=$(uname -s)

mkdir -p .build/bin
cd .build
rm -rf fabric-cli
git clone https://github.com/hyperledger/fabric-cli.git
cd fabric-cli
git checkout $FABRIC_CLI_VERSION

if [ ${envOS} = 'Darwin' ]; then
/usr/bin/sed -i '' '$a\
replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.4-0.20200626180529-18936b36feca
' go.mod
/usr/bin/sed -i '' '$a\
replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678
' go.mod
/usr/bin/sed -i '' '$a\
replace golang.org/x/sys => golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5
' go.mod
/usr/bin/sed -i '' '$a\
replace gopkg.in/check.v1 => gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
' go.mod
/usr/bin/sed -i '' '$a\
replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.3.0
' go.mod
/usr/bin/sed -i '' '$a\
replace github.com/stretchr/testify => github.com/stretchr/testify v1.5.1
' go.mod
else
sed  -e "\$areplace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.4-0.20200626180529-18936b36feca" -i go.mod
sed  -e "\$areplace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678" -i go.mod
sed  -e "\$areplace golang.org/x/sys => golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5" -i go.mod
sed  -e "\$areplace gopkg.in/check.v1 => gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127" -i go.mod
sed  -e "\$areplace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.3.0" -i go.mod
sed  -e "\$areplace github.com/stretchr/testify => github.com/stretchr/testify v1.5.1" -i go.mod
fi

make
cp ./bin/fabric ../bin/fabric
cd ../

echo "Building plugins..."

# The plugin needs to import exactly the same source code as was used to build fabric-cli, otherwise
# an error will result when the plugin is loaded. So, copy fabric-cli-ext and modify go.mod to
# replace github/hyperledger/fabric-cli with the local copy.
mkdir ./fabric-cli-ext
cp -r ../cmd/ ./fabric-cli-ext/cmd/
cp ../go.mod ./fabric-cli-ext/
cd ./fabric-cli-ext

if [ ${envOS} = 'Darwin' ]; then
/usr/bin/sed -i ''  '$a\
replace github.com/hyperledger/fabric-cli => ..\/fabric-cli' go.mod
/usr/bin/sed -i ''  '$a\
replace github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.3.0' go.mod
else
sed  -e "\$areplace github.com/hyperledger/fabric-cli => ..\/fabric-cli" -i go.mod
sed  -e "\$areplace github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.3.0" -i go.mod
fi

# ledgerconfig
go build -buildmode=plugin -o ../ledgerconfig/ledgerconfig.so ./cmd/ledgerconfig/ledgerconfig.go
cp ./cmd/ledgerconfig/plugin.yaml ../ledgerconfig/
# file
go build -buildmode=plugin -o ../file/file.so ./cmd/file/file.go
cp ./cmd/file/plugin.yaml ../file/
# extensions
go build -buildmode=plugin -o ../extensions/extensions.so ./cmd/extensions/extensions.go
cp ./cmd/extensions/plugin.yaml ../extensions/

cd ..
rm -rf ./fabric-cli
rm -rf ./fabric-cli-ext
