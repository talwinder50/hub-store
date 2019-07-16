#!/bin/sh
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

echo "Generating hub-store Test PKI"
cd /opt/go/src/github.com/trustbloc/hub-store
mkdir -p test/fixtures/keys/tls
mkdir -p test/fixtures/keys/server
mkdir -p test/fixtures/keys/client

cp /etc/ssl/openssl.cnf test/fixtures/keys/openssl.cnf
printf "\n[SAN]\nsubjectAltName=DNS:*.example.com,DNS:localhost" >> test/fixtures/keys/openssl.cnf

#create CA for TLS creds
openssl ecparam -name prime256v1 -genkey -noout -out test/fixtures/keys/tls/ec-cakey.pem
openssl req -new -x509 -key test/fixtures/keys/tls/ec-cakey.pem -subj "/C=CA/ST=ON/O=Example Internet CA TLS Inc.:CA Sec/OU=CA Sec" -out test/fixtures/keys/tls/ec-cacert.pem

#create TLS creds
openssl ecparam -name prime256v1 -genkey -noout -out test/fixtures/keys/tls/ec-key.pem
openssl req -new -key test/fixtures/keys/tls/ec-key.pem -subj "/C=CA/ST=ON/O=Example Inc.:hub-store/OU=hub-store/CN=*.example.com" -reqexts SAN -config test/fixtures/keys/openssl.cnf -out test/fixtures/keys/tls/ec-key.csr
openssl x509 -req -in test/fixtures/keys/tls/ec-key.csr -extensions SAN -extfile test/fixtures/keys/openssl.cnf -CA test/fixtures/keys/tls/ec-cacert.pem -CAkey test/fixtures/keys/tls/ec-cakey.pem -CAcreateserial -out test/fixtures/keys/tls/ec-pubCert.pem -days 365

#create CA for other creds
openssl ecparam -name prime256v1 -genkey -noout -out test/fixtures/keys/ec-cakey.pem
openssl req -new -x509 -key test/fixtures/keys/ec-cakey.pem -subj "/C=CA/ST=ON/O=Example Internet CA Inc.:CA Sec/OU=CA Sec" -out test/fixtures/keys/ec-cacert.pem

#create server creds
openssl ecparam -name prime256v1 -genkey -noout -out test/fixtures/keys/server/ec-key.pem
openssl req -new -key test/fixtures/keys/server/ec-key.pem -subj "/C=CA/ST=ON/O=Example Inc.:hub-store/OU=hub-store/CN=*.example.com" -reqexts SAN -config test/fixtures/keys/openssl.cnf -out test/fixtures/keys/server/ec-key.csr
openssl x509 -req -in test/fixtures/keys/server/ec-key.csr -extensions SAN -extfile test/fixtures/keys/openssl.cnf -CA test/fixtures/keys/ec-cacert.pem -CAkey test/fixtures/keys/ec-cakey.pem -CAcreateserial -out test/fixtures/keys/server/ec-pubCert.pem -days 365
openssl x509 -pubkey -noout -in test/fixtures/keys/server/ec-pubCert.pem > test/fixtures/keys/server/ec-pubKey.pem

#create client creds
openssl ecparam -name prime256v1 -genkey -noout -out test/fixtures/keys/client/ec-key.pem
openssl req -new -key test/fixtures/keys/client/ec-key.pem -subj "/C=CA/ST=ON/O=Example Inc.:hub-store/OU=hub-store/CN=*.example.com" -reqexts SAN -config test/fixtures/keys/openssl.cnf -out test/fixtures/keys/client/ec-key.csr
openssl x509 -req -in test/fixtures/keys/client/ec-key.csr -extensions SAN -extfile test/fixtures/keys/openssl.cnf -CA test/fixtures/keys/ec-cacert.pem -CAkey test/fixtures/keys/ec-cakey.pem -CAcreateserial -out test/fixtures/keys/client/ec-pubCert.pem -days 365

rm test/fixtures/keys/openssl.cnf
chmod -R 777 test
echo "done generating hub-store PKI"