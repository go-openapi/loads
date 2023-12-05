// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fmts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var extensions = []string{"json"}

//nolint:unparam
func assertSpecJSON(t testing.TB, specJSON []byte) bool {
	var expected map[string]interface{}
	require.NoError(t, json.Unmarshal(specJSON, &expected))

	obj := spec.Swagger{}
	require.NoError(t, json.Unmarshal(specJSON, &obj))

	cb, err := json.MarshalIndent(obj, "", "  ")
	require.NoError(t, err)

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(cb, &actual))

	return assertSpecMaps(t, actual, expected)
}

func assertSpecMaps(t testing.TB, actual, expected map[string]interface{}) bool {
	res := true
	if id, ok := expected["id"]; ok {
		res = assert.Equal(t, id, actual["id"])
	}
	res = res && assert.Equal(t, expected["consumes"], actual["consumes"])
	res = res && assert.Equal(t, expected["produces"], actual["produces"])
	res = res && assert.Equal(t, expected["schemes"], actual["schemes"])
	res = res && assert.Equal(t, expected["swagger"], actual["swagger"])
	res = res && assert.Equal(t, expected["info"], actual["info"])
	res = res && assert.Equal(t, expected["host"], actual["host"])
	res = res && assert.Equal(t, expected["basePath"], actual["basePath"])
	res = res && assert.Equal(t, expected["paths"], actual["paths"])
	res = res && assert.Equal(t, expected["definitions"], actual["definitions"])
	res = res && assert.Equal(t, expected["responses"], actual["responses"])
	res = res && assert.Equal(t, expected["securityDefinitions"], actual["securityDefinitions"])
	res = res && assert.Equal(t, expected["tags"], actual["tags"])
	res = res && assert.Equal(t, expected["externalDocs"], actual["externalDocs"])
	res = res && assert.Equal(t, expected["x-some-extension"], actual["x-some-extension"])
	res = res && assert.Equal(t, expected["x-schemes"], actual["x-schemes"])

	return res
}

//nolint:unparam
func roundTripTest(t *testing.T, fixtureType, extension, fileName string, schema interface{}) bool {
	if extension == "yaml" {
		return roundTripTestYAML(t, fixtureType, fileName, schema)
	}
	return roundTripTestJSON(t, fixtureType, fileName, schema)
}

func roundTripTestJSON(t *testing.T, fixtureType, fileName string, schema interface{}) bool {
	specName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	t.Logf("verifying %s JSON fixture %q", fixtureType, specName)

	b, err := os.ReadFile(fileName)
	require.NoError(t, err)

	var expected map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &expected))

	require.NoError(t, json.Unmarshal(b, schema))

	cb, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(cb, &actual))

	return assert.EqualValues(t, expected, actual)
}

func roundTripTestYAML(t *testing.T, fixtureType, fileName string, schema interface{}) bool {
	specName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	t.Logf("verifying %s YAML fixture %q", fixtureType, specName)

	b, err := YAMLDoc(fileName)
	require.NoError(t, err)

	var expected map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &expected))

	require.NoError(t, json.Unmarshal(b, schema))

	cb, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(cb, &actual))

	return assert.EqualValues(t, expected, actual)
}

func TestPropertyFixtures(t *testing.T) {
	for _, extension := range extensions {
		path := filepath.Join("..", "fixtures", extension, "models", "properties")
		files, err := os.ReadDir(path)
		if err != nil {
			t.Fatal(err)
		}

		// for _, f := range files {
		// 	roundTripTest(t, "property", extension, filepath.Join(path, f.Name()), &Schema{})
		// }
		f := files[0]
		roundTripTest(t, "property", extension, filepath.Join(path, f.Name()), &spec.Schema{})
	}
}

func TestAdditionalPropertiesWithObject(t *testing.T) {
	schema := new(spec.Schema)
	b, err := YAMLDoc("../fixtures/yaml/models/modelWithObjectMap.yaml")
	require.NoError(t, err)
	var expected map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &expected))
	require.NoError(t, json.Unmarshal(b, schema))

	cb, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(cb, &actual))
	assert.Equal(t, expected, actual)
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
