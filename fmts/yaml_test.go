// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package fmts

import (
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	yaml "go.yaml.in/yaml/v3"

	"github.com/go-openapi/swag/loading"
	_ "github.com/go-openapi/testify/enable/yaml/v2"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

//go:embed testdata/petstore.yaml
var yamlPetStore []byte

//go:embed testdata/with-y-key.yaml
var withYKey []byte

//go:embed testdata/with-quoted-y-key.yaml
var withQuotedYKey []byte

var errTest = errors.New("expected")

type failJSONMarshal struct{}

func (f failJSONMarshal) MarshalJSON() ([]byte, error) {
	return nil, errTest
}

func TestLoadHTTPBytes(t *testing.T) {
	_, err := loading.LoadFromFileOrHTTP("httx://12394:abd")
	require.Error(t, err)

	serv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	}))
	defer serv.Close()

	_, err = loading.LoadFromFileOrHTTP(serv.URL)
	require.Error(t, err)

	ts2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("the content"))
	}))
	defer ts2.Close()

	d, err := loading.LoadFromFileOrHTTP(ts2.URL)
	require.NoError(t, err)
	assert.Equal(t, []byte("the content"), d)
}

func TestYAMLToJSON(t *testing.T) {
	const sd = `---
1: the int key value
name: a string value
'y': some value
`
	t.Run("YAML object as JSON", func(t *testing.T) {
		var data any
		require.NoError(t, yaml.Unmarshal([]byte(sd), &data))

		d, err := YAMLToJSON(data)
		require.NoError(t, err)
		assert.JSONEqT(t,
			`{"1":"the int key value","name":"a string value","y":"some value"}`,
			string(d),
		)
	})

	t.Run("YAML nodes as JSON", func(t *testing.T) {
		var data yaml.Node
		require.NoError(t, yaml.Unmarshal([]byte(sd), &data))

		data.Content[0].Content = append(data.Content[0].Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "tag", Tag: "!!str"},
			&yaml.Node{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "name", Tag: "!!str"},
					{Kind: yaml.ScalarNode, Value: "tag name", Tag: "!!str"},
				},
			},
		)

		d, err := YAMLToJSON(data)
		require.NoError(t, err)
		assert.JSONEqT(t,
			`{"1":"the int key value","name":"a string value","y":"some value","tag":{"name":"tag name"}}`,
			string(d),
		)
	})

	t.Run("YAML slice as JSON", func(t *testing.T) {
		lst := []any{"hello"}
		d, err := YAMLToJSON(&lst)
		require.NoError(t, err)
		assert.JSONEqT(t, `["hello"]`, string(d))
	})

	t.Run("fail to convert to JSON", func(t *testing.T) {
		t.Run("with invalid receiver", func(t *testing.T) {
			_, err := YAMLToJSON(failJSONMarshal{})
			require.Error(t, err)
		})

		t.Run("with invalid document", func(t *testing.T) {
			_, err := BytesToYAMLDoc([]byte("- name: hello\n"))
			require.Error(t, err)
		})
	})

	t.Run("with BytesToYamlDoc", func(t *testing.T) {
		dd, err := BytesToYAMLDoc([]byte("description: 'object created'\n"))
		require.NoError(t, err)

		d, err := YAMLToJSON(dd)
		require.NoError(t, err)
		assert.YAMLEqT(t, `{"description":"object created"}`, string(d))
	})
}

func TestLoadStrategy(t *testing.T) {
	loader := func(_ string) ([]byte, error) {
		return yamlPetStore, nil
	}
	remLoader := func(_ string) ([]byte, error) {
		return []byte("not it"), nil
	}

	ld := loading.LoadStrategy("blah", loader, remLoader)
	b, _ := ld("")
	assert.YAMLEqT(t, string(yamlPetStore), string(b))

	serv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(yamlPetStore)
	}))
	defer serv.Close()

	s, err := YAMLDoc(serv.URL)
	require.NoError(t, err)
	assert.NotNil(t, s)

	ts2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		_, _ = rw.Write([]byte("\n"))
	}))
	defer ts2.Close()
	_, err = YAMLDoc(ts2.URL)
	require.Error(t, err)
}

func TestWithYKey(t *testing.T) {
	t.Run("with YAMLv3, unquoted y as key is parsed correctly", func(t *testing.T) {
		doc, err := BytesToYAMLDoc(withYKey)
		require.NoError(t, err)

		_, err = YAMLToJSON(doc)
		require.NoError(t, err)
	})

	t.Run("quoted y as key is parsed correctly", func(t *testing.T) {
		doc, err := BytesToYAMLDoc(withQuotedYKey)
		require.NoError(t, err)

		jsond, err := YAMLToJSON(doc)
		require.NoError(t, err)

		var yt struct {
			Definitions struct {
				Viewbox struct {
					Properties struct {
						Y struct {
							Type string `json:"type"`
						} `json:"y"`
					} `json:"properties"`
				} `json:"viewbox"`
			} `json:"definitions"`
		}
		require.NoError(t, json.Unmarshal(jsond, &yt))

		assert.EqualT(t, "integer", yt.Definitions.Viewbox.Properties.Y.Type)
	})
}
