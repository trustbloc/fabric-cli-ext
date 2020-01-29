#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

#
# Supported Targets:
#
# all:                        runs checks, unit tests, and builds the plugins
# plugins:                    builds fabric-cli and plugins
# unit-test:                  runs unit tests
# lint:                       runs linters
# checks:                     runs code checks
# generate:                   generates mocks
#

# Tool commands (overridable)
DOCKER_CMD ?= docker
GO_CMD     ?= go
ALPINE_VER ?= 3.10
GO_TAGS    ?=

# Local variables used by makefile
PROJECT_NAME            = fabric-cli-ext
ARCH                    = $(shell go env GOARCH)
GO_VER                  = $(shell grep "GO_VER" .ci-properties |cut -d'=' -f2-)
export GO111MODULE      = on
export FABRIC_CLI_VERSION ?= 97c0393410833a7ae32f2b4a540186b719642565

# Fabric tools docker image (overridable)
FABRIC_TOOLS_IMAGE   ?= hyperledger/fabric-tools
FABRIC_TOOLS_VERSION ?= 2.0.0-alpha
FABRIC_TOOLS_TAG     ?= $(ARCH)-$(FABRIC_TOOLS_VERSION)

# Fabric peer ext docker image (overridable)
FABRIC_PEER_EXT_IMAGE   ?= trustbloc/fabric-peer
FABRIC_PEER_EXT_VERSION ?= 0.1.1
FABRIC_PEER_EXT_TAG     ?= $(ARCH)-$(FABRIC_PEER_EXT_VERSION)

checks: version license lint

lint:
	@scripts/check_lint.sh

license: version
	@scripts/check_license.sh

all: clean checks unit-test plugins bddtests

unit-test:
	@scripts/unit.sh

version:
	@scripts/check_version.sh

plugins:
	@scripts/build_plugins.sh

clean: clean-images
	rm -rf ./.build
	rm -rf ./test/bddtests/fixtures/fabric/channel
	rm -rf ./test/bddtests/fixtures/fabric/crypto-config
	rm -rf ./test/bddtests/.fabriccli

generate:
	go generate ./...

crypto-gen:
	@echo "Generating crypto directory ..."
	@$(DOCKER_CMD) run -i \
		-v /$(abspath .):/opt/workspace/$(PROJECT_NAME) -u $(shell id -u):$(shell id -g) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_TAG) \
		//bin/bash -c "FABRIC_VERSION_DIR=fabric /opt/workspace/${PROJECT_NAME}/scripts/generate_crypto.sh"

channel-config-gen:
	@echo "Generating test channel configuration transactions and blocks ..."
	@$(DOCKER_CMD) run -i \
		-v /$(abspath .):/opt/workspace/$(PROJECT_NAME) -u $(shell id -u):$(shell id -g) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_TAG) \
		//bin/bash -c "FABRIC_VERSION_DIR=fabric/ /opt/workspace/${PROJECT_NAME}/scripts/generate_channeltx.sh"

populate-fixtures:
	@scripts/populate-fixtures.sh -f

bddtests: populate-fixtures docker-thirdparty bddtests-fabric-peer-docker
	@scripts/integration.sh

bddtests-fabric-peer-cli:
	@echo "Building fabric-peer cli"
	@mkdir -p ./.build/bin
	@cd test/bddtests/fixtures/fabric/peer/cmd && go build -o ../../../../../../.build/bin/fabric-peer github.com/trustbloc/fabric-cli-ext/test/bddtests/fixtures/fabric/peer/cmd

bddtests-fabric-peer-docker:
	@docker build -f ./test/bddtests/fixtures/images/fabric-peer/Dockerfile --no-cache -t trustbloc/fabric-peer-cli-test:latest \
	--build-arg FABRIC_PEER_EXT_IMAGE=$(FABRIC_PEER_EXT_IMAGE) \
	--build-arg FABRIC_PEER_EXT_TAG=$(FABRIC_PEER_EXT_TAG) \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(ALPINE_VER) \
	--build-arg GO_TAGS=$(GO_TAGS) \
	--build-arg GOPROXY=$(GOPROXY) .

docker-thirdparty:
	docker pull couchdb:2.2.0
	docker pull hyperledger/fabric-orderer:$(ARCH)-2.0.0-alpha

clean-images: CONTAINER_IDS = $(shell docker ps -a -q)
clean-images: DEV_IMAGES    = $(shell docker images dev-* -q)
clean-images:
	@echo "Stopping all containers, pruning containers and images, deleting dev images"
ifneq ($(strip $(CONTAINER_IDS)),)
	@docker stop $(CONTAINER_IDS)
endif
	@docker system prune -f
ifneq ($(strip $(DEV_IMAGES)),)
	@docker rmi $(DEV_IMAGES) -f
endif

.PHONY: all version unit-test license plugins clean clean-images generate bddtests crypto-gen channel-config-gen populate-fixtures bddtests-fabric-peer-cli bddtests-fabric-peer-docker docker-thirdparty
