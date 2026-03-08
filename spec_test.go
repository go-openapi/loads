// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package loads

import (
	_ "embed"
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

//go:embed fixtures/json/petstore-basic.json
var petStoreJSON []byte

//go:embed fixtures/yaml/search.yaml
var yamlSpec []byte

//go:embed fixtures/json/expected-expanded.json
var expectedExpanded []byte

//go:embed fixtures/json/cascade-ref-expanded.json
var cascadeRefExpanded []byte

func TestUnknownSpecVersion(t *testing.T) {
	_, err := Analyzed([]byte{}, "0.9")
	require.Error(t, err)
}

func TestDefaultsTo20(t *testing.T) {
	d, err := Analyzed(petStoreJSON, "")
	require.NoError(t, err)
	require.NotNil(t, d)

	assert.EqualT(t, "2.0", d.Version())
	assert.EqualT(t, "/api", d.BasePath())
}

func TestLoadsYAMLContent(t *testing.T) {
	d, err := Analyzed(yamlSpec, "")
	require.NoError(t, err)
	require.NotNil(t, d)

	sw := d.Spec()
	assert.EqualT(t, "1.0.0", sw.Info.Version)
}

// for issue #11.
func TestRegressionExpand(t *testing.T) {
	swaggerFile := "fixtures/yaml/swagger/1/2/3/4/swagger.yaml"
	document, err := Spec(swaggerFile)
	require.NoError(t, err)
	require.NotNil(t, document)

	d, err := document.Expanded()
	require.NoError(t, err)
	require.NotNil(t, d)

	assert.JSONMarshalAsT(t, expectedExpanded, d.Spec())
}

func TestCascadingRefExpand(t *testing.T) {
	swaggerFile := "fixtures/yaml/swagger/spec.yml"
	document, err := Spec(swaggerFile)
	require.NoError(t, err)
	require.NotNil(t, document)

	d, err := document.Expanded()
	require.NoError(t, err)
	require.NotNil(t, d)

	assert.JSONMarshalAsT(t, cascadeRefExpanded, d.Spec())
}

func TestFailsInvalidJSON(t *testing.T) {
	_, err := Analyzed(json.RawMessage([]byte("{]")), "")

	require.Error(t, err)
}

// issue go-swagger/go-swagger#1816 (regression when cloning original spec).
func TestIssue1846(t *testing.T) {
	swaggerFile := "fixtures/bugs/1816/fixture-1816.yaml"
	document, err := Spec(swaggerFile)
	require.NoError(t, err)
	require.NotNil(t, document)

	sp, err := cloneSpec(document.Spec())
	require.NoError(t, err)

	jazon, err := json.MarshalIndent(sp, "", " ")
	require.NoError(t, err)
	rex := regexp.MustCompile(`"\$ref":\s*"(.+)"`)
	m := rex.FindAllStringSubmatch(string(jazon), -1)
	require.NotNil(t, m)

	for _, matched := range m {
		subMatch := matched[1]
		require.TrueTf(t,
			strings.HasPrefix(subMatch, "#/definitions") || strings.HasPrefix(subMatch, "#/responses"),
			"expected $ref to point either to definitions or responses section, got: %s", matched[0])
	}
}

func TestEmbedded(t *testing.T) {
	swaggerFile := "fixtures/yaml/swagger/spec.yml"
	document, err := Spec(swaggerFile)
	require.NoError(t, err)
	require.NotNil(t, document)

	raw, err := json.Marshal(document.Raw())
	require.NoError(t, err)

	spc, err := json.Marshal(document.Spec())
	require.NoError(t, err)

	d, err := Embedded(raw, spc)
	require.NoError(t, err)
	require.NotNil(t, d)

	assert.JSONMarshalAsT(t, raw, d.Raw())
	assert.JSONMarshalAsT(t, spc, d.Spec())
}

func TestDocument(t *testing.T) {
	document, err := Embedded(petStoreJSON, petStoreJSON)
	require.NoError(t, err)

	require.EqualT(t, "petstore.swagger.wordnik.com", document.Host())

	require.JSONMarshalAsT(t, petStoreJSON, document.OrigSpec())
	require.JSONMarshalAsT(t, petStoreJSON, document.Pristine().Spec())

	spc := document.Spec()
	spc.Definitions = nil

	before := document.Spec()
	require.Empty(t, before.Definitions)

	reset := document.ResetDefinitions()

	require.JSONMarshalAsT(t, petStoreJSON, reset.Spec())
}

func TestSpecCircular(t *testing.T) {
	swaggerFile := "fixtures/json/resources/pathLoaderIssue.json"
	document, err := Spec(swaggerFile)
	require.NoError(t, err)
	require.NotNil(t, document)
}

func TestIssueSpec145(t *testing.T) {
	t.Run("with remote $ref", func(t *testing.T) {
		docPath := filepath.Join("fixtures", "bugs", "145", "Program Files (x86)", "AppName", "todos.json")

		t.Run("with Spec loader", func(t *testing.T) {
			document, err := Spec(docPath)
			require.NoError(t, err)
			require.NotNil(t, document)

			_, err = document.Expanded()
			require.NoError(t, err)
		})

		t.Run("with JSONSpec loader", func(t *testing.T) {
			document, err := JSONSpec(docPath)
			require.NoError(t, err)
			require.NotNil(t, document)

			_, err = document.Expanded()
			require.NoError(t, err)
		})
	})

	t.Run("with self-contained root", func(t *testing.T) {
		docPath := filepath.Join("fixtures", "bugs", "145", "Program Files (x86)", "AppName", "todos-expanded.json")

		t.Run("with Spec loader", func(t *testing.T) {
			document, err := Spec(docPath)
			require.NoError(t, err)
			require.NotNil(t, document)

			require.EqualT(t, docPath, document.SpecFilePath())

			expanded, err := document.Expanded()
			require.NoError(t, err)

			require.EqualT(t, docPath, expanded.SpecFilePath())
		})

		t.Run("with JSONSpec loader", func(t *testing.T) {
			document, err := JSONSpec(docPath)
			require.NoError(t, err)
			require.NotNil(t, document)

			_, err = document.Expanded()
			require.NoError(t, err)

			t.Run("with Pristine", func(t *testing.T) {
				pristine := document.Pristine()

				require.EqualT(t, document.SpecFilePath(), pristine.SpecFilePath())
			})
		})
	})
}
