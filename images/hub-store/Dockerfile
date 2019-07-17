#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ARG GO_VER
ARG ALPINE_VER

FROM golang:${GO_VER}-alpine${ALPINE_VER} as golang
RUN apk add --no-cache \
	gcc \
	musl-dev \
	git \
	libtool \
	bash \
	make;
ADD . $GOPATH/src/github.com/trustbloc/hub-store
WORKDIR $GOPATH/src/github.com/trustbloc/hub-store
ENV EXECUTABLES go git

FROM golang as hub-store
ARG GO_TAGS
ARG GOPROXY
RUN GO_TAGS=${GO_TAGS} GOPROXY=${GOPROXY} make hub-store


FROM alpine:${ALPINE_VER} as base
COPY --from=hub-store /go/src/github.com/trustbloc/hub-store/.build/bin/hub-store /usr/local/bin
CMD ["hub-store"]
