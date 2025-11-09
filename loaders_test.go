// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package loads

import (
	"testing"

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
