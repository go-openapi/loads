package loads_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/swag"
)

func ExampleSpec() {
	// Example with default loaders defined at the package level

	path := "fixtures/yaml/swagger/spec.yml"
	doc, err := loads.Spec(path)
	if err != nil {
		fmt.Println("Could not load this spec")
		return
	}

	fmt.Printf("Spec loaded: %q\n", doc.Host())

	// Output: Spec loaded: "api.example.com"
}

func ExampleOption() {
	// Example with custom loaders passed as options

	path := "fixtures/yaml/swagger/spec.yml"

	// a simpler version of loads.JSONDoc
	jsonLoader := loads.NewDocLoaderWithMatch(
		func(pth string) (json.RawMessage, error) {
			buf, err := ioutil.ReadFile(pth)
			return json.RawMessage(buf), err
		},
		func(pth string) bool {
			return filepath.Ext(pth) == ".json"
		},
	)

	// equivalent to the default loader at the package level, which does:
	//
	//   loads.AddLoader(swag.YAMLMatcher, swag.YAMLDoc)
	yamlLoader := loads.NewDocLoaderWithMatch(
		swag.YAMLDoc,
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
