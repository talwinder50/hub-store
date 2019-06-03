/*-
 * Copyright 2014 Square Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// file copied from github.com/square/go-jose.v2/jose-util/utils.go
// it's in a main package, copied here as a loading key utility

// ParsePublicKey loads a public key from PEM/DER/JWK-encoded data.
func ParsePublicKey(data []byte) (interface{}, error) {
	input := data

	block, _ := pem.Decode(data)
	if block != nil {
		input = block.Bytes
	}

	// Try to load SubjectPublicKeyInfo
	pub, err := x509.ParsePKIXPublicKey(input)
	if err == nil {
		return pub, nil
	}

	return nil, fmt.Errorf("square/go-jose: parse error: '%s'", err)
}

// ParsePrivateKey loads a private key from PEM/DER/JWK-encoded data.
func ParsePrivateKey(data []byte) (interface{}, error) {
	input := data

	block, _ := pem.Decode(data)
	if block != nil {
		input = block.Bytes
	}

	var priv interface{}
	priv, err := x509.ParseECPrivateKey(input)
	if err == nil {
		return priv, nil
	}

	return nil, fmt.Errorf("square/go-jose: parse error: '%s'", err)
}
