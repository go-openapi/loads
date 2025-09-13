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
	"github.com/go-openapi/swag/loading"
	"github.com/go-openapi/swag/yamlutils"
)

var (
	// YAMLMatcher matches yaml
	YAMLMatcher = loading.YAMLMatcher
	// YAMLToJSON converts YAML unmarshaled data into json compatible data
	YAMLToJSON = yamlutils.YAMLToJSON
	// BytesToYAMLDoc converts raw bytes to a map[string]interface{}
	BytesToYAMLDoc = yamlutils.BytesToYAMLDoc
	// YAMLDoc loads a yaml document from either http or a file and converts it to json
	YAMLDoc = loading.YAMLDoc
	// YAMLData loads a yaml document from either http or a file
	YAMLData = loading.YAMLData
)
