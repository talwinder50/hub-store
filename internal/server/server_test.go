/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	s := NewServer(nil)

	assert.NotNil(t, s)
}

func TestServer_GetHTTPHandler(t *testing.T) {
	s := NewServer(nil)

	req := httptest.NewRequest("POST", "/collections", nil)
	respRecorder := httptest.NewRecorder()
	handler := s.GetHTTPHandler()

	handler.ServeHTTP(respRecorder, req)

	assert.Equal(t, http.StatusNotImplemented, respRecorder.Code)
}

func TestServer_GetHTTPHandler_IncorrectMethod(t *testing.T) {
	s := NewServer(nil)

	req := httptest.NewRequest("GET", "/collections", nil)
	respRecorder := httptest.NewRecorder()
	handler := s.GetHTTPHandler()

	handler.ServeHTTP(respRecorder, req)

	assert.Equal(t, http.StatusMethodNotAllowed, respRecorder.Code)
}

func TestServer_GetHTTPHandler_IncorrectPath(t *testing.T) {
	s := NewServer(nil)

	req := httptest.NewRequest("POST", "/collection", nil)
	respRecorder := httptest.NewRecorder()
	handler := s.GetHTTPHandler()

	handler.ServeHTTP(respRecorder, req)

	assert.Equal(t, http.StatusNotFound, respRecorder.Code)
}
