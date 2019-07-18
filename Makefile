#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

#
# Supported Targets:
#
# build:              builds the binary
# unit-test:          runs unit tests
# integration-test:   runs integration tests
# lint:               runs linters
# checks:             runs build+test+lint
# generate-test-keys: generates keys for testing
# clean:              removes test fixtures
# all :               runs checks+unit-test


GO_CMD ?= go
export GO111MODULE=on

# This Makefile assumes a working Golang and Docker setup
ALPINE_VER ?= 3.9
GO_VER             = $(shell grep "GO_VER" .ci-properties |cut -d'=' -f2-)


# Namespace for the hub store
DOCKER_OUTPUT_NS          ?= trustbloc
HUB_STORE_IMAGE_NAME  ?= hub-store

#couchdb image parameters
#this couchdb image contains startup scripts that autorun and create all necessary db artifacts
export HUB_STORE_COUCHDB_IMAGE ?= trustbloc/hub-store-couchdb

export PKGS=`go list github.com/trustbloc/hub-store/... `

build:
	go build -o bin/hubstore cmd/hub-store/main.go

//TODO : Separate the couchdb test (which are dependant on external dependencies ) as integration test
unit-test: generate-test-keys
	go test -count=1 $(PKGS) -timeout=10m -coverprofile=coverage.txt -covermode=atomic ./...

license:
	@scripts/check_license.sh

lint:
	@scripts/check_lint.sh

checks: build license lint

generate-test-keys: clean
	@mkdir -p test/fixtures/keys/tls
	@docker run -i --rm \
		-v $(abspath .):/opt/go/src/github.com/trustbloc/hub-store \
		--entrypoint "/opt/go/src/github.com/trustbloc/hub-store/scripts/generate_test_keys.sh" \
		frapsoft/openssl

all: checks unit-test bddtests

hub-store:
	@echo "Building hub-store"
	@mkdir -p ./.build/bin
	@go build -o ./.build/bin/hub-store cmd/hub-store/main.go

hubstore-docker: hub-store
	@docker build -f ./images/hub-store/Dockerfile --no-cache -t $(DOCKER_OUTPUT_NS)/$(HUB_STORE_IMAGE_NAME):latest \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(ALPINE_VER) \
	--build-arg GO_TAGS=$(GO_TAGS) \
	--build-arg GOPROXY=$(GOPROXY) .

bddtests: clean checks generate-test-keys hubstore-docker
	@scripts/integration.sh

generate-test-keys: clean
	@mkdir -p test/bddtests/fixtures/keys/tls
	@docker run -i --rm \
		-v $(abspath .):/opt/go/src/github.com/trustbloc/hub-store \
		--entrypoint "/opt/go/src/github.com/trustbloc/hub-store/scripts/generate_test_keys.sh" \
		frapsoft/openssl
clean:
	rm -Rf ./.build
	rm -Rf ./test/bddtests/docker-compose.log
	rm -Rf ./test/bddtests/fixtures/keys
