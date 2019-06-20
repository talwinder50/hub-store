#!/bin/bash
#
# Copyright SecureKey Technologies Inc.
# This file contains software code that is the intellectual property of SecureKey.
# SecureKey reserves all rights in the code and you may not use it without written permission from SecureKey.
#

# This script installs dependencies for testing tools.

set -e

GO_CMD="${GO_CMD:-go}"
GOPATH="${GOPATH:-${HOME}/go}"



function installGolangCiLint {
    echo "Installing golangci-lint..."

    declare repo="github.com/golangci/golangci-lint/cmd/golangci-lint"
    declare revision="v1.15.0"
    declare pkg="github.com/golangci/golangci-lint/cmd/golangci-lint"

    installGoPkg "${repo}" "${revision}" "" "golangci-lint"
    cp -f ${BUILD_TMP}/bin/* ${GOPATH}/bin/
    rm -Rf ${GOPATH}/src/${pkg}
    mkdir -p ${GOPATH}/src/${pkg}
    cp -Rf ${BUILD_TMP}/src/${repo}/* ${GOPATH}/src/${pkg}/
}

function installGoPkg {
    declare repo=$1
    declare revision=$2
    declare pkgPath=$3
    shift 3
    declare -a cmds=$@

    echo "Installing ${repo}@${revision} to $GOPATH/bin ..."

    GOPATH=${BUILD_TMP} go get -d ${repo}
    tag=$(cd ${BUILD_TMP}/src/${repo} && git tag -l | sort -V --reverse | head -n 1 | grep "${revision}" || true)
    if [ ! -z "${tag}" ]; then
        revision=${tag}
        echo "  using tag ${revision}"
    fi
    (cd ${BUILD_TMP}/src/${repo} && git reset --hard ${revision})
    echo " Checking $GOPATH ..."
    GOPATH=${BUILD_TMP} go install -i ${repo}/${pkgPath}

    mkdir -p ${GOPATH}/bin
    for cmd in ${cmds[@]}
    do
        echo "Copying ${cmd} to ${GOPATH}/bin"
        cp -f ${BUILD_TMP}/bin/${cmd} ${GOPATH}/bin/
    done
}

function installDependencies {
    echo "Installing dependencies ..."
    export GO111MODULE=off

    BUILD_TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'didcmn'`
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/axw/gocov/...
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/AlekSi/gocov-xml
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/golang/mock/mockgen
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/client9/misspell/cmd/misspell
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u golang.org/x/tools/cmd/goimports

    installGolangCiLint

    rm -Rf ${BUILD_TMP}
}

installDependencies