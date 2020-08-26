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
export FABRIC_CLI_VERSION ?= 6d600d8656ddbb4116ee62c61658a5a5dbc87a32

# Fabric tools docker image (overridable)
FABRIC_TOOLS_IMAGE   ?= hyperledger/fabric-tools
FABRIC_TOOLS_VERSION ?= 2.0.0-alpha
FABRIC_TOOLS_TAG     ?= $(ARCH)-$(FABRIC_TOOLS_VERSION)

# Fabric peer ext docker image (overridable)
FABRIC_PEER_EXT_IMAGE   ?= trustbloc/fabric-peer
FABRIC_PEER_EXT_VERSION ?= 0.1.3
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

clean:
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

bddtests: populate-fixtures docker-thirdparty
	@scripts/integration.sh

docker-thirdparty:
	docker pull couchdb:3.1
	docker pull hyperledger/fabric-orderer:$(ARCH)-2.2.0

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
