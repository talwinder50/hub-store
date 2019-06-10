/*
   Copyright SecureKey Technologies Inc.
   This file contains software code that is the intellectual property of SecureKey.
   SecureKey reserves all rights in the code and you may not use it without written permission from SecureKey.
*/

package models

// Protected protected
type Protected struct {

	// commit strategy
	CommitStrategy string `json:"commit_strategy,omitempty"`

	// committed at
	CommittedAt string `json:"committed_at,omitempty"`

	// context
	Context string `json:"context,omitempty"`

	// interface
	Interface string `json:"interface,omitempty"`

	// kid
	Kid string `json:"kid,omitempty"`

	// meta
	Meta *ProtectedMeta `json:"meta,omitempty"`

	// object id
	ObjectID string `json:"object_id,omitempty"`

	// operation
	Operation string `json:"operation,omitempty"`

	// sub
	Sub string `json:"sub,omitempty"`

	// type
	Type string `json:"type,omitempty"`
}

// ProtectedMeta protected meta
type ProtectedMeta struct {

	// name
	Name string `json:"name,omitempty"`
}
