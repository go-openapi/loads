// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package loads

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

//go:embed fixtures/json/petstore.json
var petstoreJSON []byte

func TestLoadJSON(t *testing.T) {
	serv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(petstoreJSON)
	}))
	defer serv.Close()

	s, err := JSONSpec(serv.URL)
	require.NoError(t, err)
	require.NotNil(t, s)

	ts2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		_, _ = rw.Write([]byte("{}"))
	}))
	defer ts2.Close()
	_, err = JSONSpec(ts2.URL)
	require.Error(t, err)
}
