// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package loads_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/swag/loading"
)

// Example with default loaders defined at the package level
func ExampleSpec_file() {

	path := "fixtures/yaml/swagger/spec.yml"
	doc, err := loads.Spec(path)
	if err != nil {
		fmt.Println("Could not load this spec")
		return
	}

	fmt.Printf("Spec loaded: %q\n", doc.Host())

	// Output: Spec loaded: "api.example.com"
}

// Example with custom loaders passed as options
func ExampleLoaderOption() {
	path := "fixtures/yaml/swagger/spec.yml"

	// a simpler version of loads.JSONDoc
	jsonLoader := loads.NewDocLoaderWithMatch(
		func(pth string, _ ...loading.Option) (json.RawMessage, error) {
			buf, err := os.ReadFile(pth)
			return json.RawMessage(buf), err
		},
		func(pth string) bool {
			return filepath.Ext(pth) == ".json"
		},
	)

	// equivalent to the default loader at the package level, which does:
	//
	//   loads.AddLoader(loading.YAMLMatcher, loading.YAMLDoc)
	yamlLoader := loads.NewDocLoaderWithMatch(
		loading.YAMLDoc,
		func(pth string) bool {
			return filepath.Ext(pth) == ".yml"
		},
	)

	doc, err := loads.Spec(path, loads.WithDocLoaderMatches(jsonLoader, yamlLoader))
	if err != nil {
		fmt.Println("Could not load this spec")
		return
	}

	fmt.Printf("Spec loaded: %q\n", doc.Host())

	// Output: Spec loaded: "api.example.com"
}
