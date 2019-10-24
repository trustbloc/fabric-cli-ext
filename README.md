[![Release](https://img.shields.io/github/release/trustbloc/fabric-peer-ext.svg?style=flat-square)](https://github.com/trustbloc/fabric-cli-ext/releases/latest)

[![Build Status](https://dev.azure.com/trustbloc/fabric/_apis/build/status/trustbloc.fabric-cli-ext?branchName=master)](https://dev.azure.com/trustbloc/fabric/_build/latest?definitionId=18&branchName=master)
[![codecov](https://codecov.io/gh/trustbloc/fabric-cli-ext/branch/master/graph/badge.svg)](https://codecov.io/gh/trustbloc/fabric-cli-ext)
[![Go Report Card](https://goreportcard.com/badge/github.com/trustbloc/fabric-cli-ext?style=flat-square)](https://goreportcard.com/report/github.com/trustbloc/fabric-cli-ext)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/trustbloc/fabric-cli-ext/master/LICENSE)

# fabric-cli-ext
Command-line interface that extends Fabric's CLI.

```
git clone https://github.com/trustbloc/fabric-cli-ext.git

cd fabric-cli-ext

# run linters
make checks

# run unit-test
make unit-test

# build fabric-cli and plugins
make plugins

# install and run ledgerconfig plugin
cd .build/
bin/fabric plugin install ./ledgerconfig
bin/fabric ledgerconfig
```

## Build dependencies

* Go `1.12.x`
* Docker `18.09.x` or above

# Contributing
Thank you for your interest in contributing. Please see our [community contribution guidelines](https://github.com/trustbloc/community/blob/master/CONTRIBUTING.md) for more information.

# License
Apache License, Version 2.0 (Apache-2.0). See the [LICENSE](LICENSE) file.
