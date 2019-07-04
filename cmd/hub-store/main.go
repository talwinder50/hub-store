/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"github.com/trustbloc/hub-store/internal/server"
)

const confPageSize = "page.size"
const confServerPort = "server.port"
const confTLSCertificateFile = "tls.certificate.file"
const confTLSKeyFile = "tls.key.file"

const defaultServerPort = 8090
const defaultPageSize = 50

func main() {
	InitViperConfig()

	s := server.NewServer(getServerConfig())

	http.Handle("/hub-store", s.GetHTTPHandler())

	if err := http.ListenAndServeTLS(":"+strconv.Itoa(s.Config.Port), s.Config.TLSCertificateFile,
		s.Config.TLSKeyFile, nil); err != nil {
		fmt.Printf("Error %v\n", err)
		os.Exit(1)
	}
}

// InitViperConfig initializes the viper configuration
func InitViperConfig() {
	viper.SetDefault(confPageSize, defaultPageSize)
	viper.SetDefault(confServerPort, defaultServerPort)
	viper.SetEnvPrefix("HUB_STORE")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// getServerConfig returns the config for the server
func getServerConfig() *server.Config {
	c := server.Config{
		PageSize:           viper.GetInt(confPageSize),
		Port:               viper.GetInt(confServerPort),
		TLSCertificateFile: viper.GetString(confTLSCertificateFile),
		TLSKeyFile:         viper.GetString(confTLSKeyFile),
	}

	return &c
}
