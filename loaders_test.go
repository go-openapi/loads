package loads

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoader_EdgeCases(t *testing.T) {
	ldr := &loader{}
	ldr.Fn = JSONDoc

	// chaining with nil is a no op
	next := ldr.WithHead(nil)
	require.Equal(t, ldr, next)

	_, err := ldr.Load(`d\::invalid uri\`)
	require.Error(t, err)
}
