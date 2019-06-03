/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNewDoc(t *testing.T) {
	doc, err := NewDoc(strings.NewReader(`
		{
  			"id": "did:example:123456789abcdefghi",
			"publicKey": [{
    			"id": "did:example:123456789abcdefghi#keys-1",
    			"type": "RsaVerificationKey2018",
				"controller": "did:example:123456789abcdefghi"
  			}, {
    			"id": "did:example:123456789abcdefghi#keys-2",
    			"type": "Ed25519VerificationKey2018",
    			"controller": "did:example:pqrstuvwxyz0987654321"
  			}]
		}
	`))
	require.NoError(t, err)
	assert.Equal(t, "did:example:123456789abcdefghi", doc.ID())
	assert.Len(t, doc.Keys(), 2)
}

func TestNewDocInvalidJson(t *testing.T) {
	_, err := NewDoc(strings.NewReader("invalid"))
	require.Error(t, err)
}

func TestDocID(t *testing.T) {
	doc := &Doc{
		json: toJson(t, `{ "id": "did:example:123456789abcdefghi" }`),
	}
	assert.Equal(t, "did:example:123456789abcdefghi", doc.ID())
}

func TestDocPublicKeysLength(t *testing.T) {
	doc:= &Doc{json: toJson(t, `
		{
  			"id": "did:example:123456789abcdefghi",
			"publicKey": [{
    			"id": "did:example:123456789abcdefghi#keys-1",
    			"type": "RsaVerificationKey2018",
				"controller": "did:example:123456789abcdefghi"
  			}, {
    			"id": "did:example:123456789abcdefghi#keys-2",
    			"type": "Ed25519VerificationKey2018",
    			"controller": "did:example:pqrstuvwxyz0987654321"
  			}]
		}
	`)}
	assert.Len(t, doc.Keys(), 2)
}

func TestPublicKey(t *testing.T) {
	key := &PublicKey{json: toJson(t, `
		{
    		"id": "did:example:123456789abcdefghi#keys-1",
    		"type": "RsaVerificationKey2018",
			"controller": "did:example:123456789abcdefghi"
  		}`,
	)}
	assert.Equal(t, "did:example:123456789abcdefghi#keys-1", key.ID())
	assert.Equal(t, "RsaVerificationKey2018", key.Type())
	assert.Equal(t, "did:example:123456789abcdefghi", key.Controller())
}

func toJson(t *testing.T, raw string) map[string]interface{} {
	jsonld := make(map[string]interface{})
	if err := json.Unmarshal([]byte(raw), &jsonld); err != nil {
		require.NoError(t, err)
	}
	return jsonld
}
