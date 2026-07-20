// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package loads

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/swag/loading"
	"github.com/go-openapi/testify/v2/require"
)

func TestLoader_EdgeCases(t *testing.T) {
	ldr := &loader{}
	ldr.Fn = JSONDoc

	// chaining with nil is a no op
	next := ldr.WithHead(nil)
	require.Equal(t, ldr, next)

	_, err := ldr.Load(`d\::invalid uri\`)
	require.Error(t, err)

	clone := ldr.clone()
	cnext := clone.WithHead(nil)
	require.Equal(t, clone, cnext)
}

// errBoom is a static error used by loader tests to assert error propagation.
var errBoom = errors.New("boom")

func okLoader(tag string, sink *string) DocLoader {
	return func(_ string, _ ...loading.Option) (json.RawMessage, error) {
		if sink != nil {
			*sink = tag
		}
		return json.RawMessage(`{"loaded":"` + tag + `"}`), nil
	}
}

func TestLoaderChain(t *testing.T) {
	t.Run("empty or all-nil chain yields ErrNoLoader without panicking (never nil)", func(t *testing.T) {
		for name, chain := range map[string]DocLoader{
			"no loaders":          LoaderChain(),
			"nil Fn entry":        LoaderChain(DocLoaderWithMatch{Fn: nil}),
			"NewDocLoader nil Fn": LoaderChain(NewDocLoaderWithMatch(nil, nil)),
			"only non-matching":   LoaderChain(NewDocLoaderWithMatch(okLoader("x", nil), func(string) bool { return false })),
		} {
			t.Run(name, func(t *testing.T) {
				require.NotNil(t, chain, "LoaderChain must never return a nil DocLoader")
				require.NotPanics(t, func() {
					_, err := chain("whatever.json")
					require.ErrorIs(t, err, ErrNoLoader)
				})
			})
		}
	})

	t.Run("dispatches to the first matching loader, preserving order", func(t *testing.T) {
		var got string
		chain := LoaderChain(
			NewDocLoaderWithMatch(okLoader("yaml", &got), loading.YAMLMatcher),
			NewDocLoaderWithMatch(okLoader("json", &got), nil), // catch-all fallback
		)

		_, err := chain("spec.yaml")
		require.NoError(t, err)
		require.Equal(t, "yaml", got)

		_, err = chain("spec.json")
		require.NoError(t, err)
		require.Equal(t, "json", got)
	})

	t.Run("forwards call-time options to the matched loader", func(t *testing.T) {
		var count int
		chain := LoaderChain(NewDocLoaderWithMatch(func(_ string, opts ...loading.Option) (json.RawMessage, error) {
			count = len(opts)
			return json.RawMessage(`{}`), nil
		}, nil))

		_, err := chain("x.json", loading.WithTimeout(time.Second), loading.WithRoot(t.TempDir()))
		require.NoError(t, err)
		require.Equal(t, 2, count)
	})

	t.Run("falls through on error and aggregates as ErrLoads", func(t *testing.T) {
		chain := LoaderChain(NewDocLoaderWithMatch(func(string, ...loading.Option) (json.RawMessage, error) {
			return nil, errBoom
		}, nil))

		_, err := chain("x.json")
		require.ErrorIs(t, err, ErrLoads)
		require.ErrorIs(t, err, errBoom)
	})
}

func TestLoaderWithOptions(t *testing.T) {
	t.Run("appends fixed options after call-time options", func(t *testing.T) {
		var received int
		wrapped := LoaderWithOptions(func(_ string, opts ...loading.Option) (json.RawMessage, error) {
			received = len(opts)
			return json.RawMessage(`{}`), nil
		}, loading.WithRoot("/a"), loading.WithTimeout(time.Second))

		_, err := wrapped("x.json", loading.WithHTTPClient(nil))
		require.NoError(t, err)
		require.Equal(t, 3, received) // 1 call-time + 2 fixed
	})

	t.Run("fixed options take precedence over call-time options (last-wins)", func(t *testing.T) {
		// A confined WithRoot supplied to LoaderWithOptions must win over a call-time WithRoot
		// pointing elsewhere.
		root := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(root, "in.json"), []byte(`{"ok":true}`), 0o600))
		elsewhere := t.TempDir()

		confined := LoaderWithOptions(JSONDoc, loading.WithRoot(root))

		// call-time asks for a different root; the fixed root wins, so the in-root file loads
		b, err := confined("in.json", loading.WithRoot(elsewhere))
		require.NoError(t, err)
		require.True(t, strings.Contains(string(b), "ok"))
	})
}

// TestLoaderChain_ConfinedIntegration mirrors the go-openapi/loads consumer pattern used by
// go-swagger: a LoaderChain of option-wrapped YAML/JSON loaders confines every load, even though
// the chain itself forwards no options.
func TestLoaderChain_ConfinedIntegration(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "api.json"), []byte(`{"swagger":"2.0"}`), 0o600))
	parent := filepath.Dir(root)
	require.NoError(t, os.WriteFile(filepath.Join(parent, "secret.json"), []byte(`{"secret":true}`), 0o600))

	chain := LoaderChain(
		NewDocLoaderWithMatch(LoaderWithOptions(loading.YAMLDoc, loading.WithRoot(root)), loading.YAMLMatcher),
		NewDocLoaderWithMatch(LoaderWithOptions(loading.JSONDoc, loading.WithRoot(root)), nil),
	)

	// an in-root document loads
	b, err := chain("api.json")
	require.NoError(t, err)
	require.True(t, strings.Contains(string(b), "swagger"))

	// a path escaping the root is rejected, even though chain() passes no call-time options
	_, err = chain("../secret.json")
	require.Error(t, err)
}
