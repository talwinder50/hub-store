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
	Name string `json:"name,omitempty"`
}

// Protected gives the commit info and it is signature protected
type Protected struct {
	Interface      string    `json:"interface"`
	Context        string    `json:"context"`
	Type           string    `json:"type"`
	Operation      Operation `json:"operation"`
	CommittedAt    string    `json:"committed_at"`
	CommitStrategy string    `json:"commit_strategy"`
	Sub            string    `json:"sub"`
	Kid            string    `json:"kid"`
	ObjectID       string    `json:"object_id,omitempty"`
	Meta           *Meta     `json:"meta,omitempty"`
}

//Header defines the header parameters for Request
type Header struct {
	Revision string `json:"rev"`
	Iss      string `json:"iss"`
}

//Request defines the request structure for hub-store
type Request struct {
	Context  string `json:"@context"`
	Type     string `json:"@type"`
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	Audience string `json:"aud"`
	*CommitRequest
	*CommitQuery
	*ObjectQuery
}

// Commit gives the actual user data
type Commit struct {
	Protected string  `json:"protected"`
	Header    *Header `json:"header"`
	Payload   string  `json:"payload"`
	Signature string  `json:"signature"`
}

//CommitRequest defines the write commit struct of a request
type CommitRequest struct {
	Commit *Commit `json:"commit"`
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
	ObjectID  string   `json:"object_id"`
	Revision  []string `json:"revision"`
	SkipToken string   `json:"skip_token,omitempty"`
}

// ObjectQueryRequest defines the struct to send object query to collection store
type ObjectQueryRequest struct {
	Context   string    `json:"context"`
	Filters   []*Filter `json:"filters"`
	Interface string    `json:"interface"`
	ObjectID  []string  `json:"object_id"`
	SkipToken string    `json:"skip_token,omitempty"`
	Type      string    `json:"type"`
}

// ObjectMetadata defines the object metadata structure which will serve as response for object query
type ObjectMetadata struct {
	CommitStrategy string `json:"commit_strategy"`
	Context        string `json:"context"`
	CreatedAt      string `json:"created_at"`
	CreatedBy      string `json:"created_by"`
	ID             string `json:"id"`
	Interface      string `json:"interface"`
	Sub            string `json:"sub"`
	Type           string `json:"type"`
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
	Commits   []*Commit `json:"commits"`
	SkipToken string    `json:"skip_token,omitempty"`
}

// WriteResponse entails Base Response fields and revisions of commit along with optional skip token.
type WriteResponse struct {
	BaseResponse
	Revisions []string `json:"revisions"`
	SkipToken string   `json:"skip_token,omitempty"`
}

// ObjectQueryResponse defines the response struct for  the object query
type ObjectQueryResponse struct {
	BaseResponse
	Objects   []*ObjectMetadata `json:"objects"`
	SkipToken string            `json:"skip_token,omitempty"`
}

// BaseResponse defines the common parameters used by all different types of response.
type BaseResponse struct {
	AtContextField        string
	AtType                string
	DeveloperMessageField string
}

// Filter defines the parameters for applying filters while doing commit query on the collections store
type Filter struct {
	Field string `json:"field"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ErrorResponse defines the struct to handle errors
type ErrorResponse struct {
	DeveloperMessageField string
	ErrorCode             string `json:"error_code"`
	ErrorURL              string `json:"error_url,omitempty"`
	Target                string `json:"target"`
	UserMessage           string `json:"user_message,omitempty"`
}
