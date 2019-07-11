/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package collection

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

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
	response := &models.Response{}
	err := &models.ErrorResponse{}
	response, err = ServiceRequest(&serverConfig, request)
	require.Nil(t, err)
	assert.Equal(t, difContext, response.WriteResponse.AtContextField)
	assert.Equal(t, "WriteResponse", response.WriteResponse.AtType)
	assert.Contains(t, response.WriteResponse.Revisions, request.Commit.Header.Revision)
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

func TestCommitQuery(t *testing.T) {

	const pageSize = 1
	collectionStore := createCollectionStore()
	serverConfig := server.Config{PageSize: pageSize, CollectionStore: collectionStore}
	oid := "7e181b7ca4b04246bcc064eede4af26c"
	commitQueryReq := newCommitQueryRequest(oid, "")

	response, err := ServiceRequest(&serverConfig, commitQueryReq)
	require.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, difContext, response.CommitQueryResponse.AtContextField)
}

func TestObjectQuery(t *testing.T) {

	const pageSize = 1
	const iface = "Collections"
	const context = "http://schema.org"
	const tpe = "MusicPlaylist"
	collectionStore := createObjectQueryCollectionStore()
	serverConfig := server.Config{PageSize: pageSize, CollectionStore: collectionStore}
	oids := []string{"7e181b7ca4b04246bcc064eede4af26c", "98cccdef685843beb6412321e5b182fe", "d509499f399845faa5608a202181ea6a"}

	objectQueryReq := newObjectQueryRequest(iface, context, tpe, "", oids, []*models.Filter{})
	response, err := ServiceRequest(&serverConfig, objectQueryReq)
	require.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, difContext, response.ObjectQueryResponse.AtContextField)
	objects := response.ObjectQueryResponse.Objects
	assert.Equal(t, "did:example:123456", objects[0].CreatedBy)
	assert.Equal(t, "7e181b7ca4b04246bcc064eede4af26c", objects[0].ID)
	assert.Equal(t, "MusicPlaylist", objects[0].Type)
}
func TestObjectQueryPagination(t *testing.T) {
	const pageSize = 5
	const numObjects = 15
	require.True(t, pageSize < numObjects)

	const iface = "Collections"
	const context = "http://schema.org"
	const tpe = "MusicPlaylist"

	collectionStore := NewMockCollectionStore(nil)
	serverConfig := server.Config{PageSize: pageSize, CollectionStore: collectionStore}

	oids := make([]string, numObjects)
	objects := make(map[string]*models.Request)

	for i := 0; i < numObjects; i++ {
		req := newWriteRequest(iface, context, tpe)
		err := collectionStore.MockWriteObjectQuery(req.Commit)
		require.Nil(t, err)
		oid := req.Commit.Header.Revision
		oids = append(oids, oid)
		objects[oid] = req
	}
	objectQueryReq := newObjectQueryRequest(iface, context, tpe, "", oids, []*models.Filter{})
	response, err := ServiceRequest(&serverConfig, objectQueryReq)
	require.Nil(t, err)
	assert.NotEmpty(t, response.ObjectQueryResponse.SkipToken,
		"a skip_token was expected in the response because the pageSize is smaller than the number of written objects")
	assert.Len(t, response.Objects, pageSize, "number of results is expected to be equal to the pageSize for the first call")

	for len(response.ObjectQueryResponse.SkipToken) > 0 {
		objectQueryReq = newObjectQueryRequest(iface, context, tpe, response.ObjectQueryResponse.SkipToken, oids, []*models.Filter{})
		response, err = ServiceRequest(&serverConfig, objectQueryReq)
		require.Nil(t, err)
		if len(response.ObjectQueryResponse.SkipToken) > 0 {
			assert.Len(t, response.Objects, pageSize,
				"number of returned objects expected to be equal to pageSize if a skip_token is returned")
		} else {
			assert.True(t, len(response.Objects) <= pageSize,
				"number of returned objects expected to be equal or less than pageSie if no skip_token is returned")
		}
	}
}

func TestObjectQueryError(t *testing.T) {

	const pageSize = 1
	const iface = "invalid"
	const context = "http://schema.org"
	const tpe = "MusicPlaylist"
	collectionStore := createObjectQueryCollectionStore()
	serverConfig := server.Config{PageSize: pageSize, CollectionStore: collectionStore}
	oids := []string{"7e181b7ca4b04246bcc064eede4af26c", "98cccdef685843beb6412321e5b182fe", "d509499f399845faa5608a202181ea6a"}

	objectQueryReq := newObjectQueryRequest(iface, context, tpe, "", oids, []*models.Filter{})
	response, err := ServiceRequest(&serverConfig, objectQueryReq)
	require.NotNil(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "server_error", err.ErrorCode)
	assert.Equal(t, "failed to execute ObjectQuery against store: revision not found in the collection store", err.DeveloperMessageField)
}

func TestCommitQueryError(t *testing.T) {
	const pageSize = 1
	collectionStore := createCollectionStore()
	serverConfig := server.Config{PageSize: pageSize, CollectionStore: collectionStore}
	oid := "invalid"
	commitQueryReq := newCommitQueryRequest(oid, "")

	response, err := ServiceRequest(&serverConfig, commitQueryReq)
	require.NotNil(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "server_error", err.ErrorCode)
	assert.Equal(t, "failed to execute CommitQuery against store: revision not found in the collection store", err.DeveloperMessageField)
}
func TestDecodeProtectedError(t *testing.T) {
	input := "======"
	resp := &models.Protected{}
	err := &models.ErrorResponse{}
	resp, err = decodeProtected(input)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "illegal base64 data at input byte 0", err.DeveloperMessageField)
}
func TestDecodeProtectedUnmarshalError(t *testing.T) {
	input := "test"
	resp := &models.Protected{}
	err := &models.ErrorResponse{}
	resp, err = decodeProtected(input)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid character 'µ' looking for beginning of value", err.DeveloperMessageField)
}
func createCollectionStore() *MockCollectionStore {

	commit := &models.Commit{
		Protected: "ewoJCSAgICAiaW50ZXJmYWNlIjogIkNvbGxlY3Rpb25zIiwKCQkJImNvbnRleHQiOiAiaHR0cDovL3NjaGVtYS5vcmciLAoJCQkidHlwZSI6ICJNdXNpY1BsYXlsaXN0IiwKCQkJIm9wZXJhdGlvbiI6ICJjcmVhdGUiLAoJCQkiY29tbWl0dGVkX2F0IjogIjIwMTgtMTAtMjRUMTg6Mzk6MTAuMTA6MDBaIiwKCQkJImNvbW1pdF9zdHJhdGVneSI6ICJiYXNpYyIsCgkJCSJzdWIiOiAiZGlkOmV4YW1wbGU6YWJjMTIzIiwKCQkJImtpZCI6ICJkaWQ6ZXhhbXBsZToxMjM0NTYja2V5LWFiYyIsCgkJCSJtZXRhIjogewoJCQkibmFtZSI6ICJTYW1wbGUgcGxheWxpc3QiCgkJfQoJfQ==",
		Header: &models.Header{
			Revision: "7e181b7ca4b04246bcc064eede4af26c",
			Iss:      "did:example:123456"},
		Payload:   "ewoJCSAgICAiQGNvbnRleHQiOiAiaHR0cDovL2lkZW50aXR5LmZvdW5kYXRpb24iLAoJCQkiQHR5cGUiOiAiTXVzaWNQbGF5bGlzdCIsCgkJCSJAaWQiOiAiZm9vIiwKCQkJIm5hbWUiOiAiQSBwbGF5bGlzdCIKCX0=",
		Signature: "j3irpj90af992l"}

	collectionStore := NewMockCollectionStore(nil)
	collectionStore.Write(commit)
	return collectionStore

}
func createObjectQueryCollectionStore() *MockCollectionStore {

	commit := &models.Commit{
		Protected: "ewoJCSAgICAiaW50ZXJmYWNlIjogIkNvbGxlY3Rpb25zIiwKCQkJImNvbnRleHQiOiAiaHR0cDovL3NjaGVtYS5vcmciLAoJCQkidHlwZSI6ICJNdXNpY1BsYXlsaXN0IiwKCQkJIm9wZXJhdGlvbiI6ICJjcmVhdGUiLAoJCQkiY29tbWl0dGVkX2F0IjogIjIwMTgtMTAtMjRUMTg6Mzk6MTAuMTA6MDBaIiwKCQkJImNvbW1pdF9zdHJhdGVneSI6ICJiYXNpYyIsCgkJCSJzdWIiOiAiZGlkOmV4YW1wbGU6YWJjMTIzIiwKCQkJImtpZCI6ICJkaWQ6ZXhhbXBsZToxMjM0NTYja2V5LWFiYyIsCgkJCSJtZXRhIjogewoJCQkibmFtZSI6ICJTYW1wbGUgcGxheWxpc3QiCgkJfQoJfQ==",
		Header: &models.Header{
			Revision: "7e181b7ca4b04246bcc064eede4af26c",
			Iss:      "did:example:123456"},
		Payload:   "ewoJCSAgICAiQGNvbnRleHQiOiAiaHR0cDovL2lkZW50aXR5LmZvdW5kYXRpb24iLAoJCQkiQHR5cGUiOiAiTXVzaWNQbGF5bGlzdCIsCgkJCSJAaWQiOiAiZm9vIiwKCQkJIm5hbWUiOiAiQSBwbGF5bGlzdCIKCX0=",
		Signature: "j3irpj90af992l"}

	collectionStore := NewMockCollectionStore(nil)
	collectionStore.MockWriteObjectQuery(commit)
	return collectionStore

}
func newWriteRequest(iface, ctx, tpe string) *models.Request {

	protected := &models.Protected{
		Interface:      iface,
		Context:        ctx,
		Type:           tpe,
		Operation:      "create",
		CommittedAt:    "2018-10-24T18:39:10.10:00Z",
		CommitStrategy: "basic",
		Sub:            "did:example:abc123",
		Kid:            "did:example:123456#key-abc",
		Meta: &models.Meta{
			Name: "Sample playlist",
		},
	}
	protectedBytes, err := json.Marshal(protected)
	if err != nil {
		panic(err)
	}

	protectedString := base64.StdEncoding.EncodeToString(protectedBytes)

	payload := `{
		    "@context": "http://identity.foundation",
			"@type": "MusicPlaylist",
			"@id": "foo",
			"name": "A playlist"
	}`
	payloadString := base64.StdEncoding.EncodeToString([]byte(payload))

	req := &models.Request{
		Context:  difContext,
		Type:     "WriteRequest",
		Issuer:   "did:example:123456",
		Audience: "did:example:some-hub",
		Subject:  "did:example:abc123",
		CommitRequest: &models.CommitRequest{
			Commit: &models.Commit{
				Protected: protectedString,
				Header: &models.Header{
					Revision: randomString(),
					Iss:      "did:example:123456",
				},
				Payload:   payloadString,
				Signature: "j3irpj90af992l",
			},
		},
	}
	return req
}
func newCommitQueryRequest(oid, skipToken string) *models.Request {

	revs := []string{"7e181b7ca4b04246bcc064eede4af26c", "98cccdef685843beb6412321e5b182fe", "d509499f399845faa5608a202181ea6a"}
	req := &models.Request{
		Context:  difContext,
		Type:     "CommitQueryRequest",
		Issuer:   randomString(),
		Audience: "did:example:092u340",
		Subject:  "did:example:l2j4rlj",
		CommitQuery: &models.CommitQuery{
			CommitQueryRequest: &models.CommitQueryRequest{
				ObjectID:  oid,
				SkipToken: skipToken,
				Revision:  revs,
			}},
	}
	return req
}
func newObjectQueryRequest(iface, ctx, tpe string, skipToken string, oids []string, filters []*models.Filter) *models.Request {
	req := &models.Request{
		Context:  difContext,
		Type:     "ObjectQueryRequest",
		Issuer:   randomString(),
		Audience: "did:example:some-hub",
		Subject:  "did:example:abc123",
		ObjectQuery: &models.ObjectQuery{
			ObjectQueryRequest: &models.ObjectQueryRequest{
				Context:   ctx,
				Interface: iface,
				Type:      tpe,
				ObjectID:  oids,
				Filters:   filters,
				SkipToken: skipToken,
			},
		},
	}
	if len(skipToken) > 0 {
		req.ObjectQueryRequest.SkipToken = skipToken
	}
	return req
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

//Write mocks storing operations
func (m *MockCollectionStore) MockWriteObjectQuery(c *models.Commit) error {
	if m.Err != nil {
		return m.Err
	}
	decodedProtected, err := base64.StdEncoding.DecodeString(c.Protected)
	if err != nil {
		panic(err)
	}
	protected := &models.Protected{}
	json.Unmarshal(decodedProtected, protected)
	key := protected.Interface
	m.commit[key] = append(m.commit[key], c)
	return nil
}

//ObjectQuery mocks performing object query operations
func (m *MockCollectionStore) ObjectQuery(iface, context, tpe string, f *collection.Filter, p *db.Paging) ([]*models.Commit, string, error) {

	if m.Err != nil {
		return nil, "", m.Err
	}
	if objects, ok := m.commit[iface]; ok {
		if p.SkipToken == "" {
			paginatedObj := paginate(objects, 0, p.Size)
			skipToken := strconv.Itoa(len(objects) - p.Size)
			return paginatedObj, skipToken, nil
		} else {
			skip, err := strconv.Atoi(p.SkipToken)
			if err != nil {
				return nil, "", err
			}
			paginatedObj := paginate(objects, skip, p.Size)
			return paginatedObj, "", nil
		}

	}
	return nil, "", errors.New("revision not found in the collection store")

}

func paginate(objects []*models.Commit, skip int, size int) []*models.Commit {
	limit := func() int {
		if skip+size > len(objects) {
			return len(objects)
		} else {
			return skip + size
		}

	}

	start := func() int {
		if skip > len(objects) {
			return len(objects)
		} else {
			return skip
		}

	}
	return objects[start():limit()]
}
