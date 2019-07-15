/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package couchdb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
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

var s collection.Store

func TestMain(m *testing.M) {
	const image = "couchdb:2.3.0"
	const dbname = "test"
	if err := Pull(image, 120 * time.Second); err != nil {
		fmt.Printf("cannot pull couchdb image=%s err=%s", image, err)
		os.Exit(1)
	}
	url, cleanup := StartCouchDB(image, 30 * time.Second)
	if err := CreateDB(url, dbname, 10 * time.Second); err != nil {
		fmt.Printf("cannot create test couch database with name=%s err=%s", dbname, err)
		cleanup()
		os.Exit(1)
	}
	if err := CreateIndices(url, dbname, 10 * time.Second); err != nil {
		fmt.Printf("cannot create couchdb indices: err=%s", err)
		cleanup()
		os.Exit(1)
	}
	s = Store(&Config{URL: url, DBName: dbname})
	code := m.Run()
	cleanup()
	os.Exit(code)
}
func TestStorePanicsURL(t *testing.T) {
	cfg := &Config{URL: "", DBName: "invalid"}
	assert.Panics(t, func() { Store(cfg) }, "The code did not panic")

}
func TestStorePanicsDBName(t *testing.T) {
	cfg := &Config{URL: "invalid", DBName: ""}
	assert.Panics(t, func() { Store(cfg) }, "The code did not panic")

}
func TestParseCommitError(t *testing.T) {
	commit := &models.Commit{}
	commit.Protected = ""
	_, _, _, _, err := parseCommit(commit)
	assert.NotNil(t, err)
}
func TestParseCommitDecodedError(t *testing.T) {
	commit := &models.Commit{}
	commit.Protected = "=--=-=-="
	_, _, _, _, err := parseCommit(commit)
	assert.NotNil(t, err)
}
func TestWriteAndCommitQueryError(t *testing.T) {
	commits, _, err := s.CommitQuery("test", &collection.Filter{}, &db.Paging{SkipToken: "-1"})
	assert.Len(t, commits, 0)
	assert.Contains(t, err.Error(), "failed to execute CommitQuery on CouchDB for oid=test")
}
func TestWriteAndObjectQueryError(t *testing.T) {
	commits, _, err := s.ObjectQuery("", "", "", &collection.Filter{}, &db.Paging{SkipToken: "-1"})
	assert.Len(t, commits, 0)
	assert.Contains(t, err.Error(), "failed to execute ObjectQuery on CouchDB")
}
func TestWriteAndCommitParseError(t *testing.T) {
	commits, _, err := s.CommitQuery("test", &collection.Filter{}, &db.Paging{SkipToken: "|"})
	assert.Len(t, commits, 0)
	assert.Contains(t, err.Error(), "invalid syntax")
}
func TestWriteError(t *testing.T) {
	commit := &models.Commit{}
	commit.Protected = "invalid"
	err := s.Write(commit)
	assert.NotNil(t, s.Write(commit))
	assert.Contains(t, err.Error(), "illegal base64 data at input byte 4")
}
func TestInvalidProtectedError(t *testing.T) {
	commit1 := createWrongCommit()
	err := s.Write(commit1)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "json: cannot unmarshal string into Go value of type models.Protected")

}
func TestWriteAndObjectQuery(t *testing.T) {
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
	assert.NoError(t, s.Write(commit1))
	assert.NoError(t, s.Write(commit2))
	protected1 := decodeProtected(commit1.Protected)
	metadata, _, err := s.ObjectQuery(
		protected1.Interface, protected1.Context, protected1.Type,
		&collection.Filter{Oids: []string{commit1.Header.Revision, commit2.Header.Revision}},
		&db.Paging{},
	)
	assert.NoError(t, err)
	assert.Len(t, metadata, 2)
	assert.Contains(t, metadata, commit1)
	assert.Contains(t, metadata, commit2)
}
func TestWriteAndCommitQueryNoRevFilter(t *testing.T) {
	object := map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "MusicPlaylist",
		"@id":      "Metallica",
		"name":     "A playlist",
	}
	create := createCommit(copy(object))
	assert.NoError(t, s.Write(create))
	object["group"] = "My Favorites"
	update := updateCommit(create.Header.Revision, copy(object))
	assert.NoError(t, s.Write(update))
	commits, _, err := s.CommitQuery(create.Header.Revision, &collection.Filter{}, &db.Paging{})
	assert.NoError(t, err)
	assert.Contains(t, commits, create)
	assert.Contains(t, commits, update)
}
func TestWriteAndCommitQueryWithRevFilter(t *testing.T) {
	object := map[string]interface{}{
		"@context": "http://identity.foundation",
		"@type":    "MusicPlaylist",
		"@id":      "AC/DC",
		"name":     "A playlist",
	}
	create := createCommit(copy(object))
	assert.NoError(t, s.Write(create))
	object["group"] = "My Favorites"
	update1 := updateCommit(create.Header.Revision, copy(object))
	assert.NoError(t, s.Write(update1))
	object["genre"] = "HardRock"
	update2 := updateCommit(create.Header.Revision, copy(object))
	assert.NoError(t, s.Write(update2))
	commits, _, err := s.CommitQuery(
		create.Header.Revision,
		&collection.Filter{Revs: []string{update1.Header.Revision, update2.Header.Revision}},
		&db.Paging{})
	assert.NoError(t, err)
	assert.NotContains(t, commits, create)
	assert.Contains(t, commits, update1)
	assert.Contains(t, commits, update1)
}

func TestObjectQueryPaging(t *testing.T) {
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
		assert.NoError(t, s.Write(create))
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
	results, token, err := s.ObjectQuery(iface, context, tpe, filter, paging)
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Len(t, results, pageSize)
	for _, commit := range results {
		assert.Contains(t, objects, commit.Header.Revision)
		delete(objects, commit.Header.Revision)
	}
	for len(token) > 0 {
		paging.SkipToken = token
		results, token, err = s.ObjectQuery(iface, context, tpe, filter, paging)
		assert.NoError(t, err)
		for _, commit := range results {
			assert.Contains(t, objects, commit.Header.Revision)
			delete(objects, commit.Header.Revision)
		}
	}
	assert.Len(t, objects, 0)
}

func TestObjectQueryMetaFilter(t *testing.T) {
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
	require.NoError(t, writeAll(s, match1, match2, mismatch))
	commits, err := objectQueryAll(
		s,
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

func TestObjectQueryMetaFilterAndOidFilter(t *testing.T) {
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
	require.NoError(t, writeAll(s, object1, object2, object3))
	commits, err := objectQueryAll(
		s,
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
func TestObjectQueryWithUnsupportedFilter(t *testing.T) {
	filterType := "unsupported_type"
	filterName := "name"
	filterValue := "value"
	_, _, err := s.ObjectQuery(
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
func encodeCommitProtected(protected *models.Protected) string {
	protectedBytes, err := json.Marshal(protected)
	if err != nil {
		panic(err)
	}
	protectedEncoded := encoding(protectedBytes)

	return protectedEncoded
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
func createWrongCommit() *models.Commit {
	commit := commit()
	commit.Protected = "invalid protected"
	updatedBytes, err := json.Marshal(commit.Protected)
	if err != nil {
		panic(err)
	}
	commit.Protected = encoding(updatedBytes)
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
func encoding(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}
