// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package loads_test

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag/loading"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// These tests mutate package-level and spec-package globals; they must not run in parallel and
// restore the built-in default on cleanup.

func TestSetLoaders(t *testing.T) {
	t.Cleanup(func() { loads.SetLoaders() }) // restore the built-in default

	customLoader := func(called *bool) loads.DocLoader {
		return func(pth string, _ ...loading.Option) (json.RawMessage, error) {
			*called = true

			return loads.JSONDoc(pth)
		}
	}

	t.Run("should replace the package default chain", func(t *testing.T) {
		t.Cleanup(func() { loads.SetLoaders() })

		var called bool
		loads.SetLoaders(loads.NewDocLoaderWithMatch(customLoader(&called), nil)) // nil matcher: catch-all

		doc, err := loads.Spec("fixtures/json/petstore-basic.json")
		require.NoError(t, err)
		assert.True(t, called, "the custom loader should be used")
		assert.Equal(t, "petstore.swagger.wordnik.com", doc.Host())
	})

	t.Run("should also re-point spec.PathLoader", func(t *testing.T) {
		t.Cleanup(func() { loads.SetLoaders() })

		var called bool
		loads.SetLoaders(loads.NewDocLoaderWithMatch(customLoader(&called), nil))

		_, err := spec.PathLoader("fixtures/json/petstore-basic.json")
		require.NoError(t, err)
		assert.True(t, called, "spec.PathLoader should use the new chain")
	})

	t.Run("should restore the built-in default when called with no usable loader", func(t *testing.T) {
		loads.SetLoaders() // reset

		doc, err := loads.Spec("fixtures/yaml/swagger/spec.yml")
		require.NoError(t, err)
		assert.Equal(t, "api.example.com", doc.Host())
	})
}

func TestSetRestrictedLoaders(t *testing.T) {
	t.Cleanup(func() { loads.SetLoaders() }) // restore the built-in default

	loads.SetRestrictedLoaders("fixtures/yaml")

	t.Run("should load a spec within the root via the package default", func(t *testing.T) {
		doc, err := loads.Spec("swagger/spec.yml")
		require.NoError(t, err)
		assert.Equal(t, "api.example.com", doc.Host())
	})

	t.Run("should reject a path escaping the root", func(t *testing.T) {
		_, err := loads.Spec("../../../../etc/passwd")
		require.Error(t, err)
	})

	t.Run("should block a loopback remote URL", func(t *testing.T) {
		srv := serveSomeJSONDocument()
		defer srv.Close()

		_, err := loads.Spec(srv.URL)
		require.Error(t, err)
		assert.ErrorIs(t, err, loads.ErrForbiddenAddress)
	})

	t.Run("should confine cross-package resolution via spec.PathLoader", func(t *testing.T) {
		srv := serveSomeJSONDocument()
		defer srv.Close()

		_, err := spec.PathLoader(srv.URL)
		require.Error(t, err)
		assert.ErrorIs(t, err, loads.ErrForbiddenAddress)
	})
}
