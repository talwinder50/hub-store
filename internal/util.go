/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	"encoding/base64"
	"encoding/json"

	"github.com/trustbloc/hub-store/internal/server/models"
)

// ObjectID returns the ID of the object that this commit either creates or modifies.
// For "create" commits, the object's ID is the commit's revision (commit.header.rev).
// For "update" or "delete" commits, the object's ID is explicitly presented as commit.protected.object_id.
func ObjectID(commit *models.Commit) (string, error) {
	var oid string
	protected := models.Protected{}
	decodedBytes, err := decoding(commit.Protected)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(decodedBytes, &protected)
	if err != nil {
		return "", err
	}
	if protected.Operation == models.Create {
		oid = commit.Header.Revision
	} else {
		oid = protected.ObjectID
	}
	return oid, nil
}

func decoding(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}
