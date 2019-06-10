/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

// Doc is a DID document as defined in the Decentralized Identifiers spec:
// https://w3c-ccg.github.io/did-spec.
type Doc struct {
	json map[string]interface{}
}

// PublicKey is a key defined in the DID document's 'publicKey' attribute.
// TODO PublicKey: add support for the various key contents as per the spec (eg. publicKeyPem, publicKeyBase58, etc)
type PublicKey struct {
	json map[string]interface{}
}

// NewDoc reads the contents into a new DID document.
func NewDoc(contents io.Reader) (*Doc, error) {
	bytes, err := ioutil.ReadAll(contents)
	if err != nil {
		return nil, err
	}
	doc := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &doc); err != nil {
		return nil, err
	}
	return &Doc{json: doc}, nil
}

// ID returns the DID document's id.
func (d *Doc) ID() string {
	return toString(d.json["id"])
}

// Keys returns the DID document's set of keys.
func (d *Doc) Keys() []*PublicKey {
	keys := make([]*PublicKey, 0)
	if array, ok := d.json["publicKey"].([]interface{}); ok {
		for _, untyped := range array {
			if json, ok := untyped.(map[string]interface{}); ok {
				keys = append(keys, &PublicKey{json: json})
			}
		}
	}
	return keys
}

// ID returns the key's id.
func (pk *PublicKey) ID() string {
	return toString(pk.json["id"])
}

// Type returns the key's type.
func (pk *PublicKey) Type() string {
	return toString(pk.json["type"])
}

// Controller returns the key's controller.
func (pk *PublicKey) Controller() string {
	return toString(pk.json["controller"])
}

// Casts untyped into a string and returns it.
// Panics if it cannot be casted to a string.
func toString(untyped interface{}) string {
	if s, ok := untyped.(string); ok {
		return s
	}
	panic(fmt.Errorf("cannot cast %+v to string", untyped))
}
