// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package loads_test

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/swag/loading"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// TestIsForbiddenAddress exercises the default network policy directly, without any network:
// the dial guard's allow-path cannot be reached through httptest (which always binds loopback).
func TestIsForbiddenAddress(t *testing.T) {
	t.Run("should forbid non-public addresses", func(t *testing.T) {
		for _, s := range []string{
			"127.0.0.1", "::1", // loopback
			"::ffff:127.0.0.1",                      // IPv4-mapped loopback
			"10.0.0.1", "192.168.1.1", "172.16.0.1", // private (RFC1918)
			"fd00::1",                    // private (IPv6 ULA)
			"169.254.169.254", "fe80::1", // link-local (cloud metadata)
			"0.0.0.0", "::", // unspecified
		} {
			assert.Truef(t, loads.IsForbiddenAddress(netip.MustParseAddr(s)), "expected %s to be forbidden", s)
		}
	})

	t.Run("should allow public addresses", func(t *testing.T) {
		for _, s := range []string{
			"8.8.8.8", "1.1.1.1", "203.0.113.10", "2606:4700:4700::1111",
		} {
			assert.Falsef(t, loads.IsForbiddenAddress(netip.MustParseAddr(s)), "expected %s to be allowed", s)
		}
	})
}

func TestRestrictedHTTPClient(t *testing.T) {
	t.Run("should block a loopback destination at dial time", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
		require.NoError(t, err)

		client := loads.RestrictedHTTPClient()
		resp, err := client.Do(req) //nolint:bodyclose // the dial guard fails the request, so resp is always nil here
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, loads.ErrForbiddenAddress)
	})
}

func TestSpecRestricted(t *testing.T) {
	const root = "fixtures/yaml"

	t.Run("should load a YAML spec within the root", func(t *testing.T) {
		doc, err := loads.SpecRestricted("swagger/spec.yml", root)
		require.NoError(t, err)
		assert.Equal(t, "api.example.com", doc.Host())
	})

	t.Run("should reject a path escaping the root", func(t *testing.T) {
		_, err := loads.SpecRestricted("../../../../etc/passwd", root)
		require.Error(t, err)
	})

	t.Run("should reject an absolute path", func(t *testing.T) {
		_, err := loads.SpecRestricted("/etc/passwd", root)
		require.Error(t, err)
	})

	t.Run("should block a loopback remote URL", func(t *testing.T) {
		srv := serveSomeJSONDocument()
		defer srv.Close()

		_, err := loads.SpecRestricted(srv.URL, root)
		require.Error(t, err)
		assert.ErrorIs(t, err, loads.ErrForbiddenAddress)
	})
}

func TestJSONSpecRestricted(t *testing.T) {
	const root = "fixtures/json"

	t.Run("should load a JSON spec within the root", func(t *testing.T) {
		doc, err := loads.JSONSpecRestricted("petstore-basic.json", root)
		require.NoError(t, err)
		assert.Equal(t, "petstore.swagger.wordnik.com", doc.Host())
	})

	t.Run("should reject a path escaping the root", func(t *testing.T) {
		_, err := loads.JSONSpecRestricted("../../../../etc/passwd", root)
		require.Error(t, err)
	})

	t.Run("should block a loopback remote URL", func(t *testing.T) {
		srv := serveSomeJSONDocument()
		defer srv.Close()

		_, err := loads.JSONSpecRestricted(srv.URL, root)
		require.Error(t, err)
		assert.ErrorIs(t, err, loads.ErrForbiddenAddress)
	})
}

func TestJSONDocRestricted(t *testing.T) {
	const root = "fixtures/json"

	t.Run("should load within the root", func(t *testing.T) {
		ldr := loads.JSONDocRestricted(root)
		raw, err := ldr("petstore-basic.json")
		require.NoError(t, err)
		assert.NotEmpty(t, raw)
	})

	t.Run("should reject a path escaping the root", func(t *testing.T) {
		ldr := loads.JSONDocRestricted(root)
		_, err := ldr("../../../../etc/passwd")
		require.Error(t, err)
	})

	t.Run("should confine the loader when registered as a doc loader", func(t *testing.T) {
		doc, err := loads.Spec("petstore-basic.json", loads.WithDocLoader(loads.JSONDocRestricted(root)))
		require.NoError(t, err)
		assert.Equal(t, "petstore.swagger.wordnik.com", doc.Host())

		_, err = loads.Spec("../../../../etc/passwd", loads.WithDocLoader(loads.JSONDocRestricted(root)))
		require.Error(t, err)
	})

	t.Run("should keep confinement even when call-time options try to loosen it", func(t *testing.T) {
		// Call-time options supplied via WithLoadingOptions are honored for extras (e.g. a
		// custom root pointing elsewhere) but must not override the baked-in confinement: the
		// document still loads from the registered root.
		doc, err := loads.Spec("petstore-basic.json",
			loads.WithDocLoader(loads.JSONDocRestricted(root)),
			loads.WithLoadingOptions(loading.WithRoot("fixtures/yaml")),
		)
		require.NoError(t, err)
		assert.Equal(t, "petstore.swagger.wordnik.com", doc.Host())
	})

	t.Run("should block a loopback remote URL", func(t *testing.T) {
		srv := serveSomeJSONDocument()
		defer srv.Close()

		ldr := loads.JSONDocRestricted(root)
		_, err := ldr(srv.URL)
		require.Error(t, err)
		assert.ErrorIs(t, err, loads.ErrForbiddenAddress)
	})
}
