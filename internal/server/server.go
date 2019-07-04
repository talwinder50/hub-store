/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trustbloc/hub-store/internal/db/collection"
)

// Config is a holder for the identity hub's configuration.
type Config struct {
	// PageSize configures the page size to use for paged queries.
	PageSize int

	// Collection store.
	CollectionStore collection.Store

	// Port for the hub store
	Port int

	// TLSCertificateFile is the path to cert file
	TLSCertificateFile string

	// TLSKeyFile is the path to key file
	TLSKeyFile string
}

// Server is the server that handles http requests
type Server struct {
	// Router handles the routes for the server
	router *mux.Router

	Config *Config
}

// NewServer creates a new Server for the identity hub
func NewServer(config *Config) *Server {
	router := mux.NewRouter()
	router.HandleFunc("/collections", collectionsHandler).Methods("POST")
	router.Use(authenticationMiddleware)

	server := Server{
		router: router,
		Config: config,
	}

	return &server
}

// GetHTTPHandler returns the main http handler (router) that is used by the server
func (s *Server) GetHTTPHandler() http.Handler {
	return s.router
}

// CollectionsHandler handles the actual service request
func collectionsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: get model request and pass to collection service

	http.Error(w, "", http.StatusNotImplemented)
}

// AuthenticationMiddleware handles the JWS authentication
func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// TODO: unmarshal request into JWE msg and check auth
		next.ServeHTTP(w, r)
	})
}
