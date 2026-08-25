// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package loads

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

//go:embed testdata/json/zero-valued-bounds.json
var zeroValuedBoundsJSON []byte

// TestCloneSpecCarriesZeroValuedBounds covers what gob used to drop from the copy Analyzed takes.
//
// gob omits a struct field holding the zero value for its type and flattens a pointer to what it
// points at, so an optional number that was present and zero came back nil. "minimum": 0,
// "maximum": 0, "maxLength": 0, "maxItems": 0 and "maxProperties": 0 all disappeared, at any
// depth and with no error - from OrigSpec, which go-swagger marshals into the specification it
// embeds in a generated server, and from the definitions ResetDefinitions copies back.
func TestCloneSpecCarriesZeroValuedBounds(t *testing.T) {
	t.Parallel()

	document, err := Analyzed(json.RawMessage(zeroValuedBoundsJSON), "")
	require.NoError(t, err)

	t.Run("the copy agrees with the document it was taken from", func(t *testing.T) {
		assert.JSONMarshalAsT(t, zeroValuedBoundsJSON, document.OrigSpec())
	})

	t.Run("resetting the definitions keeps them", func(t *testing.T) {
		document.ResetDefinitions()
		assert.JSONMarshalAsT(t, zeroValuedBoundsJSON, document.Spec())
	})

	t.Run("every bound in the fixture is a zero", func(t *testing.T) {
		// guards the fixture itself: a bound respelled to a non-zero value would make the
		// test pass without exercising anything
		var raw map[string]any
		require.NoError(t, json.Unmarshal(zeroValuedBoundsJSON, &raw))

		bounds := countZeroBounds(raw)
		require.Positive(t, bounds.total, "the fixture holds no bound at all")
		assert.EqualTf(t, bounds.total, bounds.zero,
			"%d of the fixture's %d bounds are not zero", bounds.total-bounds.zero, bounds.total)
	})
}

type boundCount struct{ total, zero int }

// countZeroBounds walks raw JSON and counts the optional numbers gob elides.
func countZeroBounds(node any) boundCount {
	var count boundCount

	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			if isBoundKeyword(key) {
				number, ok := child.(float64)
				if ok {
					count.total++
					if number == 0 {
						count.zero++
					}

					continue
				}
			}

			nested := countZeroBounds(child)
			count.total += nested.total
			count.zero += nested.zero
		}
	case []any:
		for _, child := range value {
			nested := countZeroBounds(child)
			count.total += nested.total
			count.zero += nested.zero
		}
	}

	return count
}

func isBoundKeyword(key string) bool {
	switch key {
	case "minimum", "maximum", "multipleOf",
		"minLength", "maxLength",
		"minItems", "maxItems",
		"minProperties", "maxProperties":
		return true
	default:
		return strings.HasPrefix(key, "min") || strings.HasPrefix(key, "max")
	}
}
