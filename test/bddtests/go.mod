// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/hub-store/test/bddtests

go 1.11

require (
	github.com/DATA-DOG/godog v0.7.13
	github.com/fsouza/go-dockerclient v1.4.1
	github.com/go-openapi/swag v0.19.0
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	github.com/trustbloc/hub-store v0.0.0

)

replace github.com/trustbloc/hub-store => ../..
