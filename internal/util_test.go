/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trustbloc/hub-store/internal/server/models"
)

func TestObjectIdOfWithCreateCommit(t *testing.T) {
	const rev = "abc123"
	commit := &models.Commit{}
	protected := `{
			"operation": "create"
		}`
	commit.Protected = encoding([]byte(protected))
	commit.Header = &models.Header{Revision: rev}
	oid, err := ObjectID(commit)
	assert.Nil(t, err)
	assert.Equal(t, rev, oid, "ObjectID() must return commit.header.revision of create commit")
}

func TestObjectIdOfWithUpdateCommit(t *testing.T) {
	const rev = "abc123"
	commit := &models.Commit{}
	protected := &models.Protected{
		Operation: "update",
		ObjectID:  rev,
	}

	protectedBytes, err := json.Marshal(protected)
	assert.Nil(t, err)
	commit.Protected = encoding(protectedBytes)
	commit.Header = &models.Header{Revision: "999999999"}
	oid, err := ObjectID(commit)
	assert.Nil(t, err)
	assert.Equal(t, rev, oid, "ObjectID() must return commit.protected.object_id of update commit")
}
func TestObjectIdOfWithInvalidDecode(t *testing.T) {
	commit := &models.Commit{}
	protected := `{
			"test": "create"
		`
	commit.Protected = encoding([]byte(protected))
	oid, err := ObjectID(commit)
	assert.NotNil(t, err)
	assert.Equal(t, "", oid)
	assert.Equal(t, "unexpected end of JSON input", err.Error())
}
func encoding(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}
