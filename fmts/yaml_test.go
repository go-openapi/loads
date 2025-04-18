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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	yaml "gopkg.in/yaml.v3"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errTest = errors.New("expected")

type failJSONMarshal struct {
}

func (f failJSONMarshal) MarshalJSON() ([]byte, error) {
	return nil, errTest
}

func TestLoadHTTPBytes(t *testing.T) {
	_, err := swag.LoadFromFileOrHTTP("httx://12394:abd")
	require.Error(t, err)

	serv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	}))
	defer serv.Close()

	_, err = swag.LoadFromFileOrHTTP(serv.URL)
	require.Error(t, err)

	ts2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("the content"))
	}))
	defer ts2.Close()

	d, err := swag.LoadFromFileOrHTTP(ts2.URL)
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
		var data interface{}
		require.NoError(t, yaml.Unmarshal([]byte(sd), &data))

		d, err := YAMLToJSON(data)
		require.NoError(t, err)
		assert.JSONEq(t,
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
		assert.JSONEq(t,
			`{"1":"the int key value","name":"a string value","y":"some value","tag":{"name":"tag name"}}`,
			string(d),
		)
	})

	t.Run("YAML slice as JSON", func(t *testing.T) {
		lst := []interface{}{"hello"}
		d, err := YAMLToJSON(&lst)
		require.NoError(t, err)
		assert.JSONEq(t, `["hello"]`, string(d))
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
		assert.YAMLEq(t, `{"description":"object created"}`, string(d))
	})
}

func TestLoadStrategy(t *testing.T) {
	loader := func(_ string) ([]byte, error) {
		return []byte(yamlPetStore), nil
	}
	remLoader := func(_ string) ([]byte, error) {
		return []byte("not it"), nil
	}

	ld := swag.LoadStrategy("blah", loader, remLoader)
	b, _ := ld("")
	assert.YAMLEq(t, yamlPetStore, string(b))

	serv := httptest.NewServer(http.HandlerFunc(yamlPestoreServer))
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

var yamlPestoreServer = func(rw http.ResponseWriter, _ *http.Request) {
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(yamlPetStore))
}

func TestWithYKey(t *testing.T) {
	t.Run("with YAMLv3, unquoted y as key is parsed correctly", func(t *testing.T) {
		doc, err := BytesToYAMLDoc([]byte(withYKey))
		require.NoError(t, err)

		_, err = YAMLToJSON(doc)
		require.NoError(t, err)
	})

	t.Run("quoted y as key is parsed correctly", func(t *testing.T) {
		doc, err := BytesToYAMLDoc([]byte(withQuotedYKey))
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

		assert.Equal(t, "integer", yt.Definitions.Viewbox.Properties.Y.Type)
	})
}

const withQuotedYKey = `consumes:
- application/json
definitions:
  viewBox:
    type: object
    properties:
      x:
        type: integer
        format: int16
      # y -> types don't match: expect map key string or int get: bool
      "y":
        type: integer
        format: int16
      width:
        type: integer
        format: int16
      height:
        type: integer
        format: int16
info:
  description: Test RESTful APIs
  title: Test Server
  version: 1.0.0
basePath: /api
paths:
  /test:
    get:
      operationId: findAll
      parameters:
        - name: since
          in: query
          type: integer
          format: int64
        - name: limit
          in: query
          type: integer
          format: int32
          default: 20
      responses:
        200:
          description: Array[Trigger]
          schema:
            type: array
            items:
              $ref: "#/definitions/viewBox"
produces:
- application/json
schemes:
- https
swagger: "2.0"
`

const withYKey = `consumes:
- application/json
definitions:
  viewBox:
    type: object
    properties:
      x:
        type: integer
        format: int16
      # y -> types don't match: expect map key string or int get: bool
      y:
        type: integer
        format: int16
      width:
        type: integer
        format: int16
      height:
        type: integer
        format: int16
info:
  description: Test RESTful APIs
  title: Test Server
  version: 1.0.0
basePath: /api
paths:
  /test:
    get:
      operationId: findAll
      parameters:
        - name: since
          in: query
          type: integer
          format: int64
        - name: limit
          in: query
          type: integer
          format: int32
          default: 20
      responses:
        200:
          description: Array[Trigger]
          schema:
            type: array
            items:
              $ref: "#/definitions/viewBox"
produces:
- application/json
schemes:
- https
swagger: "2.0"
`

const yamlPetStore = `swagger: '2.0'
info:
  version: '1.0.0'
  title: Swagger Petstore
  description: A sample API that uses a petstore as an example to demonstrate features in the swagger-2.0 specification
  termsOfService: http://helloreverb.com/terms/
  contact:
    name: Swagger API team
    email: foo@example.com
    url: http://swagger.io
  license:
    name: MIT
    url: http://opensource.org/licenses/MIT
host: petstore.swagger.wordnik.com
basePath: /api
schemes:
  - http
consumes:
  - application/json
produces:
  - application/json
paths:
  /pets:
    get:
      description: Returns all pets from the system that the user has access to
      operationId: findPets
      produces:
        - application/json
        - application/xml
        - text/xml
        - text/html
      parameters:
        - name: tags
          in: query
          description: tags to filter by
          required: false
          type: array
          items:
            type: string
          collectionFormat: csv
        - name: limit
          in: query
          description: maximum number of results to return
          required: false
          type: integer
          format: int32
      responses:
        '200':
          description: pet response
          schema:
            type: array
            items:
              $ref: '#/definitions/pet'
        default:
          description: unexpected error
          schema:
            $ref: '#/definitions/errorModel'
    post:
      description: Creates a new pet in the store.  Duplicates are allowed
      operationId: addPet
      produces:
        - application/json
      parameters:
        - name: pet
          in: body
          description: Pet to add to the store
          required: true
          schema:
            $ref: '#/definitions/newPet'
      responses:
        '200':
          description: pet response
          schema:
            $ref: '#/definitions/pet'
        default:
          description: unexpected error
          schema:
            $ref: '#/definitions/errorModel'
  /pets/{id}:
    get:
      description: Returns a user based on a single ID, if the user does not have access to the pet
      operationId: findPetById
      produces:
        - application/json
        - application/xml
        - text/xml
        - text/html
      parameters:
        - name: id
          in: path
          description: ID of pet to fetch
          required: true
          type: integer
          format: int64
      responses:
        '200':
          description: pet response
          schema:
            $ref: '#/definitions/pet'
        default:
          description: unexpected error
          schema:
            $ref: '#/definitions/errorModel'
    delete:
      description: deletes a single pet based on the ID supplied
      operationId: deletePet
      parameters:
        - name: id
          in: path
          description: ID of pet to delete
          required: true
          type: integer
          format: int64
      responses:
        '204':
          description: pet deleted
        default:
          description: unexpected error
          schema:
            $ref: '#/definitions/errorModel'
definitions:
  pet:
    required:
      - id
      - name
    properties:
      id:
        type: integer
        format: int64
      name:
        type: string
      tag:
        type: string
  newPet:
    allOf:
      - $ref: '#/definitions/pet'
      - required:
          - name
        properties:
          id:
            type: integer
            format: int64
          name:
            type: string
  errorModel:
    required:
      - code
      - message
    properties:
      code:
        type: integer
        format: int32
      message:
        type: string
`
