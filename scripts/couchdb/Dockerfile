#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

FROM couchdb:2.3.0
LABEL maintainer=sk-dev-team

ENV COUCHDB_URL=http://localhost:5984

WORKDIR /

COPY run.sh .
COPY setup_couchdb.sh .

EXPOSE 5984

ENTRYPOINT ["/run.sh"]
