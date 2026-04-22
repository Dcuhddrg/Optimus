package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBannerContainsConfiguredAddress(t *testing.T) {
	output := Banner("v0.1.0", 8888)

	require.Contains(t, output, ":8888")
}
