// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package loads_test

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/swag/loading"
)

//go:embed fixtures
var embeddedFixtures embed.FS

// Loads a JSON document from http, with a custom header.
func ExampleJSONSpec_http_custom_header() {
	ts := serveSomeJSONDocument()
	defer ts.Close()

	doc, err := loads.JSONSpec(ts.URL,
		loads.WithLoadingOptions(loading.WithCustomHeaders(map[string]string{
			"X-API-key": "my-api-key",
		})))
	if err != nil {
		panic(err)
	}
	fmt.Println(doc.Host())
	fmt.Println(doc.Version())

	// Output:
	// api.example.com
	// 2.0
}

// Loads a YAML document and get the deserialized [spec.Swagger] specification.
func ExampleSpec_http_yaml() {
	ts := serveSomeYAMLDocument()
	defer ts.Close()

	// loads a YAML spec from a http URL
	doc, err := loads.Spec(ts.URL)
	if err != nil {
		panic(err)
	}

	fmt.Println(doc.Host())
	fmt.Println(doc.Version())

	spec := doc.Spec()
	if spec == nil {
		panic("spec should not be nil")
	}

	// Output:
	// api.example.com
	// 2.0
}

// Loads a JSON document and get the deserialized [spec.Swagger] specification.
func ExampleSpec_http_json() {
	ts := serveSomeJSONDocument()
	defer ts.Close()

	// loads a YAML spec from a http URL
	doc, err := loads.Spec(ts.URL)
	if err != nil {
		panic(err)
	}

	fmt.Println(doc.Host())
	fmt.Println(doc.Version())

	spec := doc.Spec()
	if spec == nil {
		panic("spec should not be nil")
	}

	// Output:
	// api.example.com
	// 2.0
}

// Loads a JSON document from the embedded file system and get the deserialized [spec.Swagger] specification.
func ExampleSpec_embedded_yaml() {
	// loads a YAML spec from a file on an embedded file system
	doc, err := loads.Spec(
		path.Join("fixtures", "yaml", "swagger", "spec.yml"), // [embed.FS] sep is "/" even on windows
		loads.WithLoadingOptions(
			loading.WithFS(embeddedFixtures),
		))
	if err != nil {
		panic(err)
	}

	fmt.Println(doc.Host())
	fmt.Println(doc.Version())

	spec := doc.Spec()
	if spec == nil {
		panic("spec should not be nil")
	}

	// Output:
	// api.example.com
	// 2.0
}

func serveSomeYAMLDocument() *httptest.Server {
	source, err := os.Open(filepath.Join("fixtures", "yaml", "swagger", "spec.yml"))
	if err != nil {
		panic(err)
	}

	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = io.Copy(rw, source)
	}))
}

func serveSomeJSONDocument() *httptest.Server {
	source, err := os.Open(filepath.Join("fixtures", "json", "resources", "pathLoaderIssue.json"))
	if err != nil {
		panic(err)
	}

	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = io.Copy(rw, source)
	}))
}
