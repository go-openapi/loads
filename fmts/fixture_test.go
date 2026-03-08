// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package fmts

import (
	"encoding/json"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

var extensions = []string{"json"}

//nolint:unparam // reserved for future use
func assertSpecJSON(t *testing.T, specJSON []byte) bool {
	t.Helper()

	var expected map[string]any
	require.NoError(t, json.Unmarshal(specJSON, &expected))

	obj := spec.Swagger{}
	require.NoError(t, json.Unmarshal(specJSON, &obj))

	cb, err := json.MarshalIndent(obj, "", "  ")
	require.NoError(t, err)

	var actual map[string]any
	require.NoError(t, json.Unmarshal(cb, &actual))

	return assertSpecMaps(t, actual, expected)
}

func assertSpecMaps(t *testing.T, actual, expected map[string]any) bool {
	t.Helper()

	if id, ok := expected["id"]; ok {
		if !assert.Equal(t, id, actual["id"]) {
			return false
		}
	}

	for key := range assertedKeys() {
		if !assert.Equal(t, expected[key], actual[key]) {
			return false
		}
	}

	return true
}

func assertedKeys() iter.Seq[string] {
	return slices.Values([]string{
		"consumes",
		"produces",
		"schemes",
		"swagger",
		"info",
		"host",
		"basePath",
		"paths",
		"definitions",
		"responses",
		"securityDefinitions",
		"tags",
		"externalDocs",
		"x-some-extension",
		"x-schemes",
	})
}

//nolint:unparam
func roundTripTest(t *testing.T, fixtureType, extension, fileName string, schema any) bool {
	t.Helper()

	if extension == "yaml" {
		return roundTripTestYAML(t, fixtureType, fileName, schema)
	}

	return roundTripTestJSON(t, fixtureType, fileName, schema)
}

func roundTripTestJSON(t *testing.T, fixtureType, fileName string, schema any) bool {
	t.Helper()

	specName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	t.Logf("verifying %s JSON fixture %q", fixtureType, specName)

	b, err := os.ReadFile(fileName)
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(b, schema))

	return assert.JSONMarshalAsT(t, b, schema)
}

func roundTripTestYAML(t *testing.T, fixtureType, fileName string, schema any) bool {
	t.Helper()

	specName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	t.Logf("verifying %s YAML fixture %q", fixtureType, specName)

	b, err := YAMLDoc(fileName)
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(b, schema))

	return assert.JSONMarshalAsT(t, b, schema)
}

func TestPropertyFixtures(t *testing.T) {
	for _, extension := range extensions {
		path := filepath.Join("..", "fixtures", extension, "models", "properties")
		files, err := os.ReadDir(path)
		if err != nil {
			t.Fatal(err)
		}

		f := files[0]
		roundTripTest(t, "property", extension, filepath.Join(path, f.Name()), &spec.Schema{})
	}
}

func TestAdditionalPropertiesWithObject(t *testing.T) {
	schema := new(spec.Schema)
	b, err := YAMLDoc("../fixtures/yaml/models/modelWithObjectMap.yaml")
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(b, schema))

	assert.JSONMarshalAsT(t, b, schema)
}

func TestModelFixtures(t *testing.T) {
	path := filepath.Join("..", "fixtures", "json", "models")
	files, err := os.ReadDir(path)
	require.NoError(t, err)
	specs := []string{"modelWithObjectMap", "models", "modelWithComposition", "modelWithExamples", "multipleModels"}
FILES:
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		for _, sp := range specs {
			if strings.HasPrefix(f.Name(), sp) {
				roundTripTest(t, "model", "json", filepath.Join(path, f.Name()), &spec.Schema{})
				continue FILES
			}
		}
		roundTripTest(t, "model", "json", filepath.Join(path, f.Name()), &spec.Schema{})
	}
	path = filepath.Join("..", "fixtures", "yaml", "models")
	files, err = os.ReadDir(path)
	require.NoError(t, err)

YAMLFILES:
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		for _, sp := range specs {
			if strings.HasPrefix(f.Name(), sp) {
				roundTripTest(t, "model", "yaml", filepath.Join(path, f.Name()), &spec.Schema{})
				continue YAMLFILES
			}
		}
		roundTripTest(t, "model", "yaml", filepath.Join(path, f.Name()), &spec.Schema{})
	}
}

func TestParameterFixtures(t *testing.T) {
	path := filepath.Join("..", "fixtures", "json", "resources", "parameters")
	files, err := os.ReadDir(path)
	require.NoError(t, err)

	for _, f := range files {
		roundTripTest(t, "parameter", "json", filepath.Join(path, f.Name()), &spec.Parameter{})
	}
}

func TestOperationFixtures(t *testing.T) {
	path := filepath.Join("..", "fixtures", "json", "resources", "operations")
	files, err := os.ReadDir(path)
	require.NoError(t, err)

	for _, f := range files {
		roundTripTest(t, "operation", "json", filepath.Join(path, f.Name()), &spec.Operation{})
	}
}

func TestResponseFixtures(t *testing.T) {
	path := filepath.Join("..", "fixtures", "json", "responses")
	files, err := os.ReadDir(path)
	require.NoError(t, err)

	for _, f := range files {
		if !strings.HasPrefix(f.Name(), "multiple") {
			roundTripTest(t, "response", "json", filepath.Join(path, f.Name()), &spec.Response{})
		} else {
			roundTripTest(t, "responses", "json", filepath.Join(path, f.Name()), &spec.Responses{})
		}
	}
}

func TestResourcesFixtures(t *testing.T) {
	path := filepath.Join("..", "fixtures", "json", "resources")
	files, err := os.ReadDir(path)
	require.NoError(t, err)

	pathItems := []string{"resourceWithLinkedDefinitions_part1"}
	toSkip := []string{}
FILES:
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		for _, ts := range toSkip {
			if strings.HasPrefix(f.Name(), ts) {
				t.Log("verifying resource" + strings.TrimSuffix(f.Name(), filepath.Ext(f.Name())))
				b, err := os.ReadFile(filepath.Join(path, f.Name()))
				require.NoError(t, err)
				assertSpecJSON(t, b)
				continue FILES
			}
		}
		for _, pi := range pathItems {
			if strings.HasPrefix(f.Name(), pi) {
				roundTripTest(t, "path items", "json", filepath.Join(path, f.Name()), &spec.PathItem{})
				continue FILES
			}
		}

		t.Logf("verifying resource %q", strings.TrimSuffix(f.Name(), filepath.Ext(f.Name())))
		b2, err := os.ReadFile(filepath.Join(path, f.Name()))
		require.NoError(t, err)
		assertSpecJSON(t, b2)
	}
}
