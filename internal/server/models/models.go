/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package models

// Operation is the string which represents the type of action that the user can perform on the hub-store
type Operation string

const (
	// Create is the action string used to create new objects by the user
	Create Operation = "create"
)

// Meta provides the basic info about the payload
type Meta struct {
	Name string
}

// Protected gives the commit info and it is signature protected
type Protected struct {
	Interface      string
	Context        string
	Type           string
	Operation      Operation
	CommittedAt    string
	CommitStrategy string
	Sub            string
	Kid            string
	ObjectID       string
	Meta           *Meta
}

//Header defines the header parameters for Request
type Header struct {
	Revision string
	Iss      string
}

//Request defines the request structure for hub-store
type Request struct {
	Context  string
	Type     string
	Issuer   string
	Subject  string
	Audience string
	*CommitRequest
	*CommitQuery
	*ObjectQuery
}

// Commit gives the actual user data
type Commit struct {
	Protected string
	Header    *Header
	Payload   string
	Signature string
}

//CommitRequest defines the write commit struct of a request
type CommitRequest struct {
	Commit *Commit
}

// CommitQuery defines the query struct for commit query
type CommitQuery struct {
	CommitQueryRequest *CommitQueryRequest
}

//ObjectQuery defines the query struct for object query
type ObjectQuery struct {
	ObjectQueryRequest *ObjectQueryRequest
}

// CommitQueryRequest defines the struct to send the query to the collection store
type CommitQueryRequest struct {
	ObjectID  string
	Revision  []string
	SkipToken string
}

// ObjectQueryRequest defines the struct to send object query to collection store
type ObjectQueryRequest struct {
	Context   string
	Filters   []*Filter
	Interface string
	ObjectID  []string
	SkipToken string
	Type      string
}

// ObjectMetadata defines the object metadata structure which will serve as response for object query
type ObjectMetadata struct {
	CommitStrategy string
	Context        string
	CreatedAt      string
	CreatedBy      string
	ID             string
	Interface      string
	Sub            string
	Type           string
}

// Response encapsulates different type of responses. For example: write Response CommitQuery Response etc
type Response struct {
	*WriteResponse
	*CommitQueryResponse
	*ObjectQueryResponse
}

// CommitQueryResponse commit query response
type CommitQueryResponse struct {
	BaseResponse
	Commits   []*Commit
	SkipToken string
}

// WriteResponse entails Base Response fields and revisions of commit along with optional skip token.
type WriteResponse struct {
	BaseResponse
	Revisions []string
	SkipToken string
}

// ObjectQueryResponse defines the response struct for  the object query
type ObjectQueryResponse struct {
	BaseResponse
	Objects   []*ObjectMetadata
	SkipToken string
}

// BaseResponse defines the common parameters used by all different types of response.
type BaseResponse struct {
	AtContextField        string
	AtType                string
	DeveloperMessageField string
}

// Filter defines the parameters for applying filters while doing commit query on the collections store
type Filter struct {
	Field string
	Type  string
	Value string
}

// ErrorResponse defines the struct to handle errors
type ErrorResponse struct {
	DeveloperMessageField string
	ErrorCode             string
	ErrorURL              string
	Target                string
	UserMessage           string
}
