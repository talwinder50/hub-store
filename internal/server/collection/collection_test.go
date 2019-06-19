/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package collection

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/trustbloc/hub-store/internal/db"
	"github.com/trustbloc/hub-store/internal/db/collection"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/trustbloc/hub-store/internal/server"
	"github.com/trustbloc/hub-store/internal/server/models"
)

// standard @context
var difContext = "https://schema.identity.foundation/0.1"

const pageSize = 1

// Hub Store must return a WriteResponse that includes the model's known revisions.
func TestWriteRequest(t *testing.T) {
	collectionStore := NewMockCollectionStore(nil)
	serverConfig := server.Config{PageSize: pageSize, CollectionStore: collectionStore}
	request := newWriteRequest("Collections", "http://schema.org", "MusicPlaylist")
	writeResponse := &models.Response{}
	err := &models.ErrorResponse{}
	writeResponse, err = ServiceRequest(&serverConfig, request)
    require.Nil(t, err)
	assert.Equal(t, difContext, writeResponse.AtContextField)
	assert.Equal(t, "WriteResponse", writeResponse.AtType)
	assert.Contains(t, writeResponse.Revisions, request.Commit.Header.Revision)
}

func TestInvalidWriteRequest(t *testing.T) {
	collectionStore := NewMockCollectionStore(nil)
	serverConfig := server.Config{PageSize: pageSize, CollectionStore: collectionStore}
	request := &models.Request{Type: "InvalidRequest"}
	writeResponse, err := ServiceRequest(&serverConfig, request)
	assert.Nil(t, writeResponse)
	assert.Equal(t, err.ErrorCode, "bad_request")
	assert.Equal(t, err.DeveloperMessageField, "unsupported request type")
}

// Hub Store must return a standard ErrorResponse with correct headers and error_code when WriteRequest fails internally.
func TestWriteRequestError(t *testing.T) {
	testErr := errors.New("mock store not available")
	collectionStore := NewMockCollectionStore(testErr)
	serverConfig := server.Config{PageSize: pageSize, CollectionStore: collectionStore}
	request := newWriteRequest("Collections", "http://schema.org", "MusicPlaylist")
	writeResponse, err := ServiceRequest(&serverConfig, request)
	assert.Nil(t, writeResponse)
	assert.NotNil(t, err)
	assert.Equal(t, "server_error", err.ErrorCode)
	assert.Equal(t, "failed to commit to store: mock store not available", err.DeveloperMessageField)
}

func newWriteRequest(iface, ctx, tpe string) *models.Request {
	req := &models.Request{}

	protected := fmt.Sprintf(`{
		    "interface": "%s",
			"context": "%s",
			"type": "%s",
			"operation": "create",
			"committed_at": "2018-10-24T18:39:10.10:00Z",
			"commit_strategy": "basic",
			"sub": "did:example:abc123",
			"kid": "did:example:123456#key-abc",
			"meta": {
			"name": "Sample playlist"
		}
	}`, iface, ctx, tpe)

	payload := `{
		    "@context": "http://identity.foundation",
			"@type": "MusicPlaylist",
			"@id": "foo",
			"name": "A playlist"
	}`
	//Request commit has encoded protected and payload
	err := json.NewDecoder(strings.NewReader(fmt.Sprintf(
		`{
            "@context": "https://schema.identity.foundation/0.1",
            "@type": "WriteRequest",
            "iss": "did:example:123456",
            "aud": "did:example:some-hub",
            "sub": "did:example:abc123",
            "commit": {
				"protected": "%s",
                "header": {
                    "rev": "%s",
                    "iss": "did:example:123456"
                },
                "payload": "%s",
                "signature": "j3irpj90af992l"
            }
        }`, encoding([]byte(protected)), randomString(), encoding([]byte(payload)),
	))).Decode(req)
	if err != nil {
		panic(fmt.Sprintf("failed to decode test request: %s", err))
	}
	return req
}
func encoding(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}
func randomString() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}

// MockCollectionStore  mocks collection.Store. for testing purposes.
type MockCollectionStore struct {
	// Commit cache.
	commit map[string][]*models.Commit
	Err    error
}

// NewMockCollectionStore creates mock for collection store
func NewMockCollectionStore(err error) *MockCollectionStore {
	return &MockCollectionStore{commit: make(map[string][]*models.Commit), Err: err}
}

//Write mocks storing operations
func (m *MockCollectionStore) Write(c *models.Commit) error {
	if m.Err != nil {
		return m.Err
	}
	keyRevision := c.Header.Revision
	m.commit[keyRevision] = append(m.commit[keyRevision], c)
	return nil
}

//CommitQuery mocks committing query operations
func (m *MockCollectionStore) CommitQuery(oid string, f *collection.Filter, p *db.Paging) ([]*models.Commit, string, error) {
	if m.Err != nil {
		return nil, "", m.Err
	}
	if commits, ok := m.commit[oid]; ok {
		return commits, "", nil

	}
	return nil, "", errors.New("revision not found in the collection store")
}
