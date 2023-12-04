package loads

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const optionFixture = "fixtures/json/resources/pathLoaderIssue.json"

func TestOptionsWithDocLoader(t *testing.T) {
	document, err := Spec(optionFixture, WithDocLoader(func(pth string) (json.RawMessage, error) {
		buf, err := os.ReadFile(pth)
		return json.RawMessage(buf), err
	}))
	require.NoError(t, err)
	require.NotNil(t, document)
	require.NotNil(t, document.pathLoader)

	b, err := document.pathLoader.Load(optionFixture)
	require.NoError(t, err)

	trimmed, err := trimData(b)
	require.NoError(t, err)

	assert.EqualValues(t, trimmed, document.Raw())

	// a nil loader is a no op
	_, err = Spec(optionFixture, WithDocLoader(nil))
	require.NoError(t, err)
}

func TestOptionsLoaderFromOptions(t *testing.T) {
	var called int

	// not chaining here, just replacing with the last one
	l := loaderFromOptions([]LoaderOption{
		WithDocLoader(func(pth string) (json.RawMessage, error) {
			called = 1
			buf, err := os.ReadFile(pth)
			return json.RawMessage(buf), err
		}),
		WithDocLoader(func(pth string) (json.RawMessage, error) {
			called = 2
			buf, err := os.ReadFile(pth)
			return json.RawMessage(buf), err
		}),
	})
	require.NotNil(t, l)

	b, err := l.Load(optionFixture)
	require.NoError(t, err)
	require.NotNil(t, b)

	require.Equal(t, 2, called)
}

func TestOptionsWithDocLoaderMatches(t *testing.T) {
	jsonLoader := NewDocLoaderWithMatch(
		func(pth string) (json.RawMessage, error) {
			buf, err := os.ReadFile(pth)
			return json.RawMessage(buf), err
		},
		func(pth string) bool {
			return filepath.Ext(pth) == ".json"
		},
	)

	document, err := Spec(optionFixture, WithDocLoaderMatches(jsonLoader))
	require.NoError(t, err)
	require.NotNil(t, document)
	require.NotNil(t, document.pathLoader)

	yamlLoader := NewDocLoaderWithMatch(
		swag.YAMLDoc,
		func(pth string) bool {
			return filepath.Ext(pth) == ".yaml"
		},
	)

	document, err = Spec(optionFixture, WithDocLoaderMatches(yamlLoader))
	require.Error(t, err)
	require.Nil(t, document)

	// chained loaders, with different ordering
	document, err = Spec(optionFixture, WithDocLoaderMatches(yamlLoader, jsonLoader))
	require.NoError(t, err)
	require.NotNil(t, document)

	document, err = Spec(optionFixture, WithDocLoaderMatches(jsonLoader, yamlLoader))
	require.NoError(t, err)
	require.NotNil(t, document)

	// the nil loader is a no op
	nilLoader := NewDocLoaderWithMatch(nil, nil)
	document, err = Spec(optionFixture, WithDocLoaderMatches(nilLoader, jsonLoader, yamlLoader))
	require.NoError(t, err)
	require.NotNil(t, document)

	// the nil matcher always matches
	nilMatcher := NewDocLoaderWithMatch(func(_ string) (json.RawMessage, error) {
		return nil, errors.New("test")
	}, nil)
	_, err = Spec(optionFixture, WithDocLoaderMatches(nilMatcher))
	require.Error(t, err)
	require.Equal(t, "test", err.Error())

	// when a matcher returns an errors, the next one is tried
	document, err = Spec(optionFixture, WithDocLoaderMatches(nilMatcher, jsonLoader, yamlLoader))
	require.NoError(t, err)
	require.NotNil(t, document)
}
