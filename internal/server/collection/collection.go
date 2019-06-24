/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package collection

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/trustbloc/hub-store/internal"
	"github.com/trustbloc/hub-store/internal/db"
	"github.com/trustbloc/hub-store/internal/db/collection"
	"github.com/trustbloc/hub-store/internal/server"
	"github.com/trustbloc/hub-store/internal/server/models"
)

//ServiceRequest handles the filtered request according to its request type
func ServiceRequest(config *server.Config, request *models.Request) (*models.Response, *models.ErrorResponse) {
	switch request.Type {
	case "WriteRequest":
		return doWrite(config, request.Commit)
	case "CommitQueryRequest":
		return doCommitQuery(config, request.Query)
	default:
		return nil, badRequest("unsupported request type", request.Type)
	}
}

func doWrite(config *server.Config, commit *models.Commit) (*models.Response, *models.ErrorResponse) {
	if err := config.CollectionStore.Write(commit); err != nil {
		msg := fmt.Sprintf("failed to commit to store: %s", err)
		log.Error(msg)
		return nil, serverError(msg)
	}
	oid, err := internal.ObjectID(commit)
	if err != nil {
		return nil, serverError(err.Error())
	}
	commits, _, err := config.CollectionStore.CommitQuery(
		oid,
		&collection.Filter{},
		&db.Paging{},
	)
	if err != nil {
		msg := fmt.Sprintf("failed to query commits from store: %s", err)
		log.Error(msg)
		return nil, serverError(msg)
	}
	revs := make([]string, len(commits))
	for i, c := range commits {
		revs[i] = c.Header.Revision
	}
	wr := writeResponse(revs)
	response := &models.Response{WriteResponse: wr}
	return response, nil
}

func doCommitQuery(config *server.Config, req *models.CommitQueryRequest) (*models.Response, *models.ErrorResponse) {
	commits, token, err := config.CollectionStore.CommitQuery(
		req.ObjectID,
		&collection.Filter{Revs: req.Revision},
		&db.Paging{
			Size:      config.PageSize,
			SkipToken: req.SkipToken,
		},
	)
	if err != nil {
		msg := fmt.Sprintf("failed to execute CommitQuery against store: %s", err)
		log.Error(msg)
		return nil, serverError(msg)
	}
	cq := commitQueryResponse(commits, token)
	response := &models.Response{CommitQueryResponse: cq}
	return response, nil
}

// "bad_request" error response
func badRequest(msg string, target string) *models.ErrorResponse {
	return errResponse(msg, "bad_request", target)
}

// "server_error" error response
func serverError(msg string) *models.ErrorResponse {
	return errResponse(msg, "server_error", "")
}

func writeResponse(revs []string) *models.WriteResponse {
	var context = "https://schema.identity.foundation/0.1"
	return &models.WriteResponse{
		BaseResponse: models.BaseResponse{
			AtContextField:        context,
			AtType:                "WriteResponse",
			DeveloperMessageField: ""},
		Revisions: revs}
}

func commitQueryResponse(commits []*models.Commit, token string) *models.CommitQueryResponse {
	var context = "https://schema.identity.foundation/0.1"
	return &models.CommitQueryResponse{
		BaseResponse: models.BaseResponse{
			AtContextField:        context,
			AtType:                "CommitQueryResponse",
			DeveloperMessageField: ""},
		Commits:   commits,
		SkipToken: token}
}

func errResponse(msg, errCode, target string) *models.ErrorResponse {
	er := &models.ErrorResponse{}
	er.ErrorCode = errCode
	er.DeveloperMessageField = msg
	er.Target = target
	return er
}
