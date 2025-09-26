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

package loads

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnknownSpecVersion(t *testing.T) {
	_, err := Analyzed([]byte{}, "0.9")
	require.Error(t, err)
}

func TestDefaultsTo20(t *testing.T) {
	d, err := Analyzed(PetStoreJSONMessage, "")
	require.NoError(t, err)
	require.NotNil(t, d)

	assert.Equal(t, "2.0", d.Version())
	// assert.Equal(t, "2.0", d.data["swagger"].(string))
	assert.Equal(t, "/api", d.BasePath())
}

func TestLoadsYAMLContent(t *testing.T) {
	d, err := Analyzed(json.RawMessage([]byte(YAMLSpec)), "")
	require.NoError(t, err)
	require.NotNil(t, d)

	sw := d.Spec()
	assert.Equal(t, "1.0.0", sw.Info.Version)
}

// for issue 11
func TestRegressionExpand(t *testing.T) {
	swaggerFile := "fixtures/yaml/swagger/1/2/3/4/swagger.yaml"
	document, err := Spec(swaggerFile)
	require.NoError(t, err)
	require.NotNil(t, document)

	d, err := document.Expanded()
	require.NoError(t, err)
	require.NotNil(t, d)

	b, _ := d.Spec().MarshalJSON()
	assert.JSONEq(t, expectedExpanded, string(b))
}

func TestCascadingRefExpand(t *testing.T) {
	swaggerFile := "fixtures/yaml/swagger/spec.yml"
	document, err := Spec(swaggerFile)
	require.NoError(t, err)
	require.NotNil(t, document)

	d, err := document.Expanded()
	require.NoError(t, err)
	require.NotNil(t, d)

	b, _ := d.Spec().MarshalJSON()
	assert.JSONEq(t, cascadeRefExpanded, string(b))
}

func TestFailsInvalidJSON(t *testing.T) {
	_, err := Analyzed(json.RawMessage([]byte("{]")), "")

	require.Error(t, err)
}

// issue go-swagger/go-swagger#1816 (regression when cloning original spec)
func TestIssue1846(t *testing.T) {
	swaggerFile := "fixtures/bugs/1816/fixture-1816.yaml"
	document, err := Spec(swaggerFile)
	require.NoError(t, err)
	require.NotNil(t, document)

	sp, err := cloneSpec(document.Spec())
	require.NoError(t, err)

	jazon, _ := json.MarshalIndent(sp, "", " ")
	rex := regexp.MustCompile(`"\$ref":\s*"(.+)"`)
	m := rex.FindAllStringSubmatch(string(jazon), -1)
	require.NotNil(t, m)

	for _, matched := range m {
		subMatch := matched[1]
		require.Truef(t,
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

	rawEmbedded, err := json.Marshal(d.Raw())
	require.NoError(t, err)

	spcEmbedded, err := json.Marshal(d.Spec())
	require.NoError(t, err)

	assert.JSONEq(t, string(raw), string(rawEmbedded))
	assert.JSONEq(t, string(spc), string(spcEmbedded))
}

func TestDocument(t *testing.T) {
	document, err := Embedded(PetStoreJSONMessage, PetStoreJSONMessage)
	require.NoError(t, err)

	require.Equal(t, "petstore.swagger.wordnik.com", document.Host())

	orig, err := json.Marshal(document.OrigSpec())
	require.NoError(t, err)

	require.JSONEq(t, string(PetStoreJSONMessage), string(orig))

	cloned, err := json.Marshal(document.Pristine().Spec())
	require.NoError(t, err)

	require.JSONEq(t, string(PetStoreJSONMessage), string(cloned))

	spc := document.Spec()
	spc.Definitions = nil

	before := document.Spec()
	require.Empty(t, before.Definitions)

	reset := document.ResetDefinitions()

	afterReset, err := json.Marshal(reset.Spec())
	require.NoError(t, err)

	require.JSONEq(t, string(PetStoreJSONMessage), string(afterReset))
}

func BenchmarkAnalyzed(b *testing.B) {
	d := []byte(`{
  "swagger": "2.0",
  "info": {
    "version": "1.0.0",
    "title": "Swagger Petstore",
    "contact": {
      "name": "Wordnik API Team",
      "url": "http://developer.wordnik.com"
    },
    "license": {
      "name": "Creative Commons 4.0 International",
      "url": "http://creativecommons.org/licenses/by/4.0/"
    }
  },
  "host": "petstore.swagger.wordnik.com",
  "basePath": "/api",
  "schemes": [
    "http"
  ],
  "paths": {
    "/pets": {
      "get": {
        "security": [
          {
            "basic": []
          }
        ],
        "tags": [ "Pet Operations" ],
        "operationId": "getAllPets",
        "parameters": [
          {
            "name": "status",
            "in": "query",
            "description": "The status to filter by",
            "type": "string"
          },
          {
            "name": "limit",
            "in": "query",
            "description": "The maximum number of results to return",
            "type": "integer",
						"format": "int64"
          }
        ],
        "summary": "Finds all pets in the system",
        "responses": {
          "200": {
            "description": "Pet response",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Pet"
              }
            }
          },
          "default": {
            "description": "Unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "post": {
        "security": [
          {
            "basic": []
          }
        ],
        "tags": [ "Pet Operations" ],
        "operationId": "createPet",
        "summary": "Creates a new pet",
        "consumes": ["application/x-yaml"],
        "produces": ["application/x-yaml"],
        "parameters": [
          {
            "name": "pet",
            "in": "body",
            "description": "The Pet to create",
            "required": true,
            "schema": {
              "$ref": "#/definitions/newPet"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Created Pet response",
            "schema": {
              "$ref": "#/definitions/Pet"
            }
          },
          "default": {
            "description": "Unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      }
    }`)

	for i := range 1000 {
		d = append(d, []byte(`,
    "/pets/`)...)
		d = strconv.AppendInt(d, int64(i), 10)
		d = append(d, []byte(`": {
      "delete": {
        "security": [
          {
            "apiKey": []
          }
        ],
        "description": "Deletes the Pet by id",
        "operationId": "deletePet",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "description": "ID of pet to delete",
            "required": true,
            "type": "integer",
            "format": "int64"
          }
        ],
        "responses": {
          "204": {
            "description": "pet deleted"
          },
          "default": {
            "description": "unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "get": {
        "tags": [ "Pet Operations" ],
        "operationId": "getPetById",
        "summary": "Finds the pet by id",
        "responses": {
          "200": {
            "description": "Pet response",
            "schema": {
              "$ref": "#/definitions/Pet"
            }
          },
          "default": {
            "description": "Unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "parameters": [
        {
          "name": "id",
          "in": "path",
          "description": "ID of pet",
          "required": true,
          "type": "integer",
          "format": "int64"
        }
      ]
    }`)...)
	}

	d = append(d, []byte(`
  },
  "definitions": {
    "Category": {
      "id": "Category",
      "properties": {
        "id": {
          "format": "int64",
          "type": "integer"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "Pet": {
      "id": "Pet",
      "properties": {
        "category": {
          "$ref": "#/definitions/Category"
        },
        "id": {
          "description": "unique identifier for the pet",
          "format": "int64",
          "maximum": 100.0,
          "minimum": 0.0,
          "type": "integer"
        },
        "name": {
          "type": "string"
        },
        "photoUrls": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "status": {
          "description": "pet status in the store",
          "enum": [
            "available",
            "pending",
            "sold"
          ],
          "type": "string"
        },
        "tags": {
          "items": {
            "$ref": "#/definitions/Tag"
          },
          "type": "array"
        }
      },
      "required": [
        "id",
        "name"
      ]
    },
    "newPet": {
      "anyOf": [
        {
          "$ref": "#/definitions/Pet"
        },
        {
          "required": [
            "name"
          ]
        }
      ]
    },
    "Tag": {
      "id": "Tag",
      "properties": {
        "id": {
          "format": "int64",
          "type": "integer"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "Error": {
      "required": [
        "code",
        "message"
      ],
      "properties": {
        "code": {
          "type": "integer",
          "format": "int32"
        },
        "message": {
          "type": "string"
        }
      }
    }
  },
  "consumes": [
    "application/json",
    "application/xml"
  ],
  "produces": [
    "application/json",
    "application/xml",
    "text/plain",
    "text/html"
  ],
  "securityDefinitions": {
    "basic": {
      "type": "basic"
    },
    "apiKey": {
      "type": "apiKey",
      "in": "header",
      "name": "X-API-KEY"
    }
  }
}
`)...)
	rm := json.RawMessage(d)
	b.ResetTimer()
	for b.Loop() {
		_, err := Analyzed(rm, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

const YAMLSpec = `swagger: '2.0'

info:
  version: "1.0.0"
  title: Simple Search API
  description: |
    A very simple api description that makes a x-www-form-urlencoded only API to submit searches.

produces:
  - application/json

consumes:
  - application/json

paths:
  /search:
    post:
      operationId: search
      summary: searches tasks
      description: searches the task titles and descriptions for a match
      consumes:
        - application/x-www-form-urlencoded
      parameters:
        - name: q
          in: formData
          type: string
          description: the search string
          required: true
  /tasks:
    get:
      operationId: getTasks
      summary: Gets Task objects.
      description: |
        Optional query param of **size** determines
        size of returned array
      tags:
        - tasks
      parameters:
        - name: size
          in: query
          description: Size of task list
          type: integer
          format: int32
          default: 20
        - name: completed
          in: query
          description: when true shows completed tasks
          type: boolean

      responses:
        default:
          description: Generic Error
        200:
          description: Successful response
          headers:
            X-Rate-Limit:
              type: integer
              format: int32
            X-Rate-Limit-Remaining:
              type: integer
              format: int32
              default: 42
            X-Rate-Limit-Reset:
              type: integer
              format: int32
              default: "1449875311"
            X-Rate-Limit-Reset-Human:
              type: string
              default: 3 days
            X-Rate-Limit-Reset-Human-Number:
              type: string
              default: 3
            Access-Control-Allow-Origin:
              type: string
              default: "*"
          schema:
            type: array
            items:
              $ref: "#/definitions/Task"
    post:
      operationId: createTask
      summary: Creates a 'Task' object.
      description: |
        Validates the content property for length etc.
      parameters:
        - name: body
          in: body
          schema:
            $ref: "#/definitions/Task"
      tags:
        - tasks
      responses:
        default:
          description: Generic Error
        201:
          description: Task Created

  /tasks/{id}:
    parameters:
      - name: id
        in: path
        type: integer
        format: int32
        description: The id of the task
        required: true
        minimum: 1
    put:
      operationId: updateTask
      summary: updates a task.
      description: |
        Validates the content property for length etc.
      tags:
        - tasks
      parameters:
        - name: body
          in: body
          description: the updated task
          schema:
            $ref: "#/definitions/Task"
      responses:
        default:
          description: Generic Error
        200:
          description: Task updated
          schema:
            $ref: "#/definitions/Task"
    delete:
      operationId: deleteTask
      summary: deletes a task
      description: |
        Deleting a task is irrevocable.
      tags:
        - tasks
      responses:
        default:
          description: Generic Error
        204:
          description: Task Deleted


definitions:
  Task:
    title: A Task object
    description: |
      This describes a task. Tasks require a content property to be set.
    required:
      - content
    type: object
    properties:
      id:
        title: the unique id of the task
        description: |
          This id property is autogenerated when a task is created.
        type: integer
        format: int64
        readOnly: true
      content:
        title: The content of the task
        description: |
          Task content can contain [GFM](https://help.github.com/articles/github-flavored-markdown/).
        type: string
        minLength: 5
      completed:
        title: when true this task is completed
        type: boolean
      creditcard:
        title: the credit card format usage
        type: string
        format: creditcard
      createdAt:
        title: task creation time
        type: string
        format: date-time
        readOnly: true
`

// PetStoreJSONMessage json raw message for Petstore20
var PetStoreJSONMessage = json.RawMessage([]byte(PetStore20))

// PetStore20 json doc for swagger 2.0 pet store
const PetStore20 = `{
  "swagger": "2.0",
  "info": {
    "version": "1.0.0",
    "title": "Swagger Petstore",
    "contact": {
      "name": "Wordnik API Team",
      "url": "http://developer.wordnik.com"
    },
    "license": {
      "name": "Creative Commons 4.0 International",
      "url": "http://creativecommons.org/licenses/by/4.0/"
    }
  },
  "host": "petstore.swagger.wordnik.com",
  "basePath": "/api",
  "schemes": [
    "http"
  ],
  "paths": {
    "/pets": {
      "get": {
        "security": [
          {
            "basic": []
          }
        ],
        "tags": [ "Pet Operations" ],
        "operationId": "getAllPets",
        "parameters": [
          {
            "name": "status",
            "in": "query",
            "description": "The status to filter by",
            "type": "string"
          },
          {
            "name": "limit",
            "in": "query",
            "description": "The maximum number of results to return",
            "type": "integer",
						"format": "int64"
          }
        ],
        "summary": "Finds all pets in the system",
        "responses": {
          "200": {
            "description": "Pet response",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Pet"
              }
            }
          },
          "default": {
            "description": "Unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "post": {
        "security": [
          {
            "basic": []
          }
        ],
        "tags": [ "Pet Operations" ],
        "operationId": "createPet",
        "summary": "Creates a new pet",
        "consumes": ["application/x-yaml"],
        "produces": ["application/x-yaml"],
        "parameters": [
          {
            "name": "pet",
            "in": "body",
            "description": "The Pet to create",
            "required": true,
            "schema": {
              "$ref": "#/definitions/newPet"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Created Pet response",
            "schema": {
              "$ref": "#/definitions/Pet"
            }
          },
          "default": {
            "description": "Unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      }
    },
    "/pets/{id}": {
      "delete": {
        "security": [
          {
            "apiKey": []
          }
        ],
        "description": "Deletes the Pet by id",
        "operationId": "deletePet",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "description": "ID of pet to delete",
            "required": true,
            "type": "integer",
            "format": "int64"
          }
        ],
        "responses": {
          "204": {
            "description": "pet deleted"
          },
          "default": {
            "description": "unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "get": {
        "tags": [ "Pet Operations" ],
        "operationId": "getPetById",
        "summary": "Finds the pet by id",
        "responses": {
          "200": {
            "description": "Pet response",
            "schema": {
              "$ref": "#/definitions/Pet"
            }
          },
          "default": {
            "description": "Unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "parameters": [
        {
          "name": "id",
          "in": "path",
          "description": "ID of pet",
          "required": true,
          "type": "integer",
          "format": "int64"
        }
      ]
    }
  },
  "definitions": {
    "Category": {
      "id": "Category",
      "properties": {
        "id": {
          "format": "int64",
          "type": "integer"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "Pet": {
      "id": "Pet",
      "properties": {
        "category": {
          "$ref": "#/definitions/Category"
        },
        "id": {
          "description": "unique identifier for the pet",
          "format": "int64",
          "maximum": 100.0,
          "minimum": 0.0,
          "type": "integer"
        },
        "name": {
          "type": "string"
        },
        "photoUrls": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "status": {
          "description": "pet status in the store",
          "enum": [
            "available",
            "pending",
            "sold"
          ],
          "type": "string"
        },
        "tags": {
          "items": {
            "$ref": "#/definitions/Tag"
          },
          "type": "array"
        }
      },
      "required": [
        "id",
        "name"
      ]
    },
    "newPet": {
      "anyOf": [
        {
          "$ref": "#/definitions/Pet"
        },
        {
          "required": [
            "name"
          ]
        }
      ]
    },
    "Tag": {
      "id": "Tag",
      "properties": {
        "id": {
          "format": "int64",
          "type": "integer"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "Error": {
      "required": [
        "code",
        "message"
      ],
      "properties": {
        "code": {
          "type": "integer",
          "format": "int32"
        },
        "message": {
          "type": "string"
        }
      }
    }
  },
  "consumes": [
    "application/json",
    "application/xml"
  ],
  "produces": [
    "application/json",
    "application/xml",
    "text/plain",
    "text/html"
  ],
  "securityDefinitions": {
    "basic": {
      "type": "basic"
    },
    "apiKey": {
      "type": "apiKey",
      "in": "header",
      "name": "X-API-KEY"
    }
  }
}
`

const expectedExpanded = `
{
   "produces":[
      "application/json",
      "plain/text"
   ],
   "schemes":[
      "https",
      "http"
   ],
   "swagger":"2.0",
   "info":{
      "description":"Something",
      "title":"Something",
      "contact":{
         "name":"Somebody",
         "url":"https://url.com",
         "email":"email@url.com"
      },
      "version":"v1"
   },
   "host":"security.sonusnet.com",
   "basePath":"/api",
   "paths":{
      "/whatnot":{
         "get":{
            "description":"Get something",
            "responses":{
               "200":{
                  "description":"The something",
                  "schema":{
                     "description":"A collection of service events",
                     "type":"object",
                     "properties":{
                        "page":{
                           "description":"A description of a paged result",
                           "type":"object",
                           "properties":{
                              "page":{
                                 "description":"the page that was requested",
                                 "type":"integer"
                              },
                              "page_items":{
                                 "description":"the number of items per page requested",
                                 "type":"integer"
                              },
                              "pages":{
                                 "description":"the total number of pages available",
                                 "type":"integer"
                              },
                              "total_items":{
                                 "description":"the total number of items available",
                                 "type":"integer",
                                 "format":"int64"
                              }
                           }
                        },
                        "something":{
                           "description":"Something",
                           "type":"object",
                           "properties":{
                              "p1":{
                                 "description":"A string",
                                 "type":"string"
                              },
                              "p2":{
                                 "description":"An integer",
                                 "type":"integer"
                              }
                           }
                        }
                     }
                  }
               },
               "500":{
                  "description":"Oops"
               }
            }
         }
      }
   },
   "definitions":{
      "Something":{
         "description":"A collection of service events",
         "type":"object",
         "properties":{
            "page":{
               "description":"A description of a paged result",
               "type":"object",
               "properties":{
                  "page":{
                     "description":"the page that was requested",
                     "type":"integer"
                  },
                  "page_items":{
                     "description":"the number of items per page requested",
                     "type":"integer"
                  },
                  "pages":{
                     "description":"the total number of pages available",
                     "type":"integer"
                  },
                  "total_items":{
                     "description":"the total number of items available",
                     "type":"integer",
                     "format":"int64"
                  }
               }
            },
            "something":{
               "description":"Something",
               "type":"object",
               "properties":{
                  "p1":{
                     "description":"A string",
                     "type":"string"
                  },
                  "p2":{
                     "description":"An integer",
                     "type":"integer"
                  }
               }
            }
         }
      }
   }
}
`

const cascadeRefExpanded = `
{
  "swagger": "2.0",
  "consumes":[
     "application/json"
  ],
  "produces":[
     "application/json"
  ],
  "schemes":[
     "http"
  ],
	"host": "api.example.com",
  "info":{
     "description":"recursively following JSON references",
     "title":"test 1",
     "contact":{
        "name":"Fred"
     },
     "version":"0.1.1"
  },
  "paths":{
     "/getAll":{
        "get":{
           "operationId":"getAll",
           "parameters":[
              {
                 "description":"max number of results",
                 "name":"a",
                 "in":"body",
                 "schema":{
                    "type":"string"
                 }
              }
           ],
           "responses":{
              "200":{
                 "description":"Success",
                 "schema":{
                    "type":"array",
                    "items":{
                       "type":"string"
                    }
                 }
              }
           }
        }
     }
  },
  "definitions":{
     "a":{
        "type":"string"
     },
     "b":{
        "type":"array",
        "items":{
           "type":"string"
        }
     }
  }
}
`

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

			require.Equal(t, docPath, document.SpecFilePath())

			expanded, err := document.Expanded()
			require.NoError(t, err)

			require.Equal(t, docPath, expanded.SpecFilePath())
		})

		t.Run("with JSONSpec loader", func(t *testing.T) {
			document, err := JSONSpec(docPath)
			require.NoError(t, err)
			require.NotNil(t, document)

			_, err = document.Expanded()
			require.NoError(t, err)

			t.Run("with Pristine", func(t *testing.T) {
				pristine := document.Pristine()

				require.Equal(t, document.SpecFilePath(), pristine.SpecFilePath())
			})
		})
	})
}
