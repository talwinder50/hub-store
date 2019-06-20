/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package commontests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/hub-store/internal/db/collection"

	"github.com/trustbloc/hub-store/internal/db"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/trustbloc/hub-store/internal/server/models"
)

// RunTestsFor tests the functionalities like writing an object, querying an object with and without paging and
// querying based on filters.
func RunTestsFor(t *testing.T, s collection.Store) {
	testWriteAndObjectQuery(t, s)
	testWriteAndCommitQueryNoRevFilter(t, s)
	testWriteAndCommitQueryWithRevFilter(t, s)
	testObjectQueryPaging(t, s)
	testObjectQueryMetaFilter(t, s)
	testObjectQueryMetaFilterAndOidFilter(t, s)
	testObjectQueryWithUnsupportedFilter(t, s)
}
func testWriteAndObjectQuery(t *testing.T, store collection.Store) {
	commit1 := createCommit(
		map[string]interface{}{
			"@context": "http://identity.foundation",
			"@type":    "MusicPlaylist",
			"@id":      "Alice in Chains",
			"name":     "A playlist",
		})
	commit2 := createCommit(
		map[string]interface{}{
			"@context": "http://identity.foundation",
			"@type":    "MusicPlaylist",
			"@id":      "Guns n Roses",
			"name":     "A playlist",
		})
	assert.NoError(t, store.Write(commit1))
	assert.NoError(t, store.Write(commit2))
	protected1 := decodeProtected(commit1.Protected)
	metadata, _, err := store.ObjectQuery(
		protected1.Interface, protected1.Context, protected1.Type,
		&collection.Filter{Oids: []string{commit1.Header.Revision, commit2.Header.Revision}},
		&db.Paging{},
	)
	assert.NoError(t, err)
	assert.Len(t, metadata, 2)
	assert.Contains(t, metadata, commit1)
	assert.Contains(t, metadata, commit2)
}
func testWriteAndCommitQueryNoRevFilter(t *testing.T, store collection.Store) {
	object := map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "MusicPlaylist",
		"@id":      "Metallica",
		"name":     "A playlist",
	}
	create := createCommit(copy(object))
	assert.NoError(t, store.Write(create))
	object["group"] = "My Favorites"
	update := updateCommit(create.Header.Revision, copy(object))
	assert.NoError(t, store.Write(update))
	commits, _, err := store.CommitQuery(create.Header.Revision, &collection.Filter{}, &db.Paging{})
	assert.NoError(t, err)
	assert.Contains(t, commits, create)
	assert.Contains(t, commits, update)
}

func testWriteAndCommitQueryWithRevFilter(t *testing.T, store collection.Store) {
	object := map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "MusicPlaylist",
		"@id":      "AC/DC",
		"name":     "A playlist",
	}
	create := createCommit(copy(object))
	assert.NoError(t, store.Write(create))
	object["group"] = "My Favorites"
	update1 := updateCommit(create.Header.Revision, copy(object))
	assert.NoError(t, store.Write(update1))
	object["genre"] = "HardRock"
	update2 := updateCommit(create.Header.Revision, copy(object))
	assert.NoError(t, store.Write(update2))
	commits, _, err := store.CommitQuery(
		create.Header.Revision,
		&collection.Filter{Revs: []string{update1.Header.Revision, update2.Header.Revision}},
		&db.Paging{})
	assert.NoError(t, err)
	assert.NotContains(t, commits, create)
	assert.Contains(t, commits, update1)
	assert.Contains(t, commits, update1)
}

func testObjectQueryPaging(t *testing.T, store collection.Store) {
	const maxCommits = 20
	const pageSize = 7
	object := map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "MusicPlaylist",
		"@id":      "Queen",
	}
	oids := make([]string, 0)
	objects := make(map[string]*models.Commit)
	for i := 0; i < maxCommits; i++ {
		create := createCommit(object)
		objects[create.Header.Revision] = create
		oids = append(oids, create.Header.Revision)
		assert.NoError(t, store.Write(create))
	}
	decodedBytes, err := decoding(objects[oids[0]].Protected)
	assert.NoError(t, err)
	ObjProtected := &models.Protected{}
	err = json.Unmarshal(decodedBytes, ObjProtected)
	assert.NoError(t, err)

	iface := ObjProtected.Interface
	context := ObjProtected.Context
	tpe := ObjProtected.Type
	filter := &collection.Filter{Oids: oids}
	paging := &db.Paging{Size: pageSize}
	results, token, err := store.ObjectQuery(iface, context, tpe, filter, paging)
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Len(t, results, pageSize)
	for _, commit := range results {
		assert.Contains(t, objects, commit.Header.Revision)
		delete(objects, commit.Header.Revision)
	}
	for len(token) > 0 {
		paging.SkipToken = token
		results, token, err = store.ObjectQuery(iface, context, tpe, filter, paging)
		assert.NoError(t, err)
		for _, commit := range results {
			assert.Contains(t, objects, commit.Header.Revision)
			delete(objects, commit.Header.Revision)
		}
	}
	assert.Len(t, objects, 0)
}

func testObjectQueryMetaFilter(t *testing.T, store collection.Store) {
	match1 := createCommit(map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "Sports",
		"@id":      "Baseball",
	})
	match1Protected := decodeProtected(match1.Protected)
	match1Protected.Meta.Name = "the_right_name"
	match1.Protected = encodeCommitProtected(match1Protected)

	match2 := createCommit(map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "Sports",
		"@id":      "Basketball",
	})
	match2Protected := decodeProtected(match2.Protected)

	match2Protected.Meta.Name = match1Protected.Meta.Name
	match2.Protected = encodeCommitProtected(match2Protected)

	mismatch := createCommit(map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "Dunno",
		"@id":      "OddOneOut",
	})
	mismatchProtected := decodeProtected(mismatch.Protected)
	mismatchProtected.Meta.Name = randomString()
	requireEqualInterfaceContextType(t, match1, match2, mismatch)
	require.Equal(t, match1Protected.Meta.Name, match2Protected.Meta.Name)
	require.NotEqual(t, match1Protected.Meta.Name, mismatchProtected.Meta.Name)
	require.NoError(t, writeAll(store, match1, match2, mismatch))
	commits, err := objectQueryAll(
		store,
		match1Protected.Interface, match1Protected.Context, match1Protected.Type,
		&collection.Filter{
			MetadataFilters: []*models.Filter{newEqualsFilter("name", match1Protected.Meta.Name)},
		},
	)
	require.NoError(t, err)
	assert.Len(t, commits, 2)
	assert.Contains(t, commits, match1)
	assert.Contains(t, commits, match2)
}

func testObjectQueryMetaFilterAndOidFilter(t *testing.T, store collection.Store) {
	object1 := createCommit(map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "Sports",
		"@id":      "Baseball",
	})
	object1Protected := decodeProtected(object1.Protected)
	object1Protected.Meta.Name = "the_right_name"

	object1.Protected = encodeCommitProtected(object1Protected)

	object2 := createCommit(map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "Sports",
		"@id":      "Basketball",
	})

	object2Protected := decodeProtected(object2.Protected)
	object2Protected.Meta.Name = object1Protected.Meta.Name

	object2.Protected = encodeCommitProtected(object2Protected)

	object3 := createCommit(map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "Dunno",
		"@id":      "OddOneOut",
	})

	object3Protected := decodeProtected(object3.Protected)

	object3Protected.Meta.Name = randomString()

	object3.Protected = encodeCommitProtected(object3Protected)

	requireEqualInterfaceContextType(t, object1, object2, object3)
	require.Equal(t, object1Protected.Meta.Name, object2Protected.Meta.Name)
	require.NotEqual(t, object1Protected.Meta.Name, object3Protected.Meta.Name)
	require.NoError(t, writeAll(store, object1, object2, object3))
	commits, err := objectQueryAll(
		store,
		object1Protected.Interface, object1Protected.Context, object1Protected.Type,
		&collection.Filter{
			Oids:            []string{object1.Header.Revision},
			MetadataFilters: []*models.Filter{newEqualsFilter("name", object1Protected.Meta.Name)},
		},
	)
	require.NoError(t, err)
	assert.Len(t, commits, 1)
	assert.Contains(t, commits, object1)
}

func encodeCommitProtected(protected *models.Protected) string {
	protectedBytes, err := json.Marshal(protected)
	if err != nil {
		panic(err)
	}
	protectedEncoded := encoding(protectedBytes)

	return protectedEncoded
}
func testObjectQueryWithUnsupportedFilter(t *testing.T, store collection.Store) {
	filterType := "unsupported_type"
	filterName := "name"
	filterValue := "value"
	_, _, err := store.ObjectQuery(
		"Collections", "http://schema.org", "MusicPlaylist",
		&collection.Filter{
			MetadataFilters: []*models.Filter{{
				Type:  filterType,
				Field: filterName,
				Value: filterValue,
			}},
		},
		&db.Paging{},
	)
	assert.Error(t, err)
	assert.IsType(t, &collection.ErrUnsupportedFilter{}, err)
}

func writeAll(store collection.Store, commits ...*models.Commit) error {
	for _, c := range commits {
		if err := store.Write(c); err != nil {
			return err
		}
	}
	return nil
}

func objectQueryAll(store collection.Store, iface, context, tpe string, filter *collection.Filter) (commits []*models.Commit, err error) {
	var token string
	paging := &db.Paging{}
	commits, token, err = store.ObjectQuery(iface, context, tpe, filter, paging)
	if err != nil {
		return nil, err
	}
	for len(token) > 0 {
		paging.SkipToken = token
		var moreCommits []*models.Commit
		moreCommits, token, err = store.ObjectQuery(iface, context, tpe, filter, paging)
		if err != nil {
			return nil, err
		}
		commits = append(commits, moreCommits...)
	}
	return commits, err
}

func newEqualsFilter(field, value string) *models.Filter {
	op := "eq"
	return &models.Filter{
		Type:  op,
		Field: field,
		Value: value,
	}
}

func requireEqualInterfaceContextType(t *testing.T, commits ...*models.Commit) {
	if len(commits) < 2 {
		panic("you should have provided at least two commits to this function in your test!")
	}
	reference := commits[0]

	decodedBytes, err := decoding(reference.Protected)
	assert.NoError(t, err)
	referenceProtected := &models.Protected{}
	err = json.Unmarshal(decodedBytes, referenceProtected)
	assert.NoError(t, err)

	for _, c := range commits[1:] {
		decodedBytes, err := decoding(c.Protected)
		assert.NoError(t, err)
		cProtected := &models.Protected{}
		err = json.Unmarshal(decodedBytes, cProtected)
		assert.NoError(t, err)
		require.Equal(t, referenceProtected.Interface, cProtected.Interface)
		require.Equal(t, referenceProtected.Context, cProtected.Context)
		require.Equal(t, referenceProtected.Type, cProtected.Type)
	}
}
func createCommit(payload map[string]interface{}) *models.Commit {
	commit := commit()
	commitProtected := decodeProtected(commit.Protected)
	commitProtected.Operation = "create"
	updatedBytes, err := json.Marshal(commitProtected)
	if err != nil {
		panic(err)
	}
	commit.Protected = encoding(updatedBytes)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	encodedPayload := encoding(payloadBytes)
	commit.Payload = encodedPayload
	return commit
}

func updateCommit(objID string, payload map[string]interface{}) *models.Commit {
	commit := commit()
	commitProtected := decodeProtected(commit.Protected)
	commitProtected.Operation = "update"
	commitProtected.ObjectID = objID
	updatedBytes, err := json.Marshal(commitProtected)
	if err != nil {
		panic(err)
	}
	commit.Protected = encoding(updatedBytes)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	encodedPayload := encoding(payloadBytes)
	commit.Payload = encodedPayload
	return commit
}

func commit() *models.Commit {
	protected := fmt.Sprintf(`{
		    "interface": "Collections",
			"context": "http://schema.org",
			"type": "MusicPlaylist",
			"committed_at": "%s",
			"commit_strategy": "basic",
			"sub": "did:example:abc123",
			"kid": "did:example:123456#key-abc",
			"meta": {
			"name": "Sample playlist"
		}
	}`, time.Now().String())
	return &models.Commit{
		Protected: encoding([]byte(protected)),
		Header: &models.Header{
			Revision: randomString(),
			Iss:      "did:example:123456",
		},
	}
}
func randomString() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}

func copy(object map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for key, val := range object {
		copy[key] = val
	}
	return copy
}

func decodeProtected(protected string) *models.Protected {
	decodedBytes, err := decoding(protected)
	if err != nil {
		panic(err)
	}
	protected1 := &models.Protected{}
	err = json.Unmarshal(decodedBytes, protected1)
	if err != nil {
		panic(err)
	}
	return protected1
}
func decoding(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}
func encoding(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}
