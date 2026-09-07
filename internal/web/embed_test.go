package web

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetsIncludesIndex(t *testing.T) {
	assets, err := Assets()
	require.NoError(t, err)

	_, err = fs.ReadFile(assets, "index.html")
	require.NoError(t, err)
}

func TestFallbackAssetsIncludePlaceholderIndex(t *testing.T) {
	fallback, err := fs.Sub(assetFS, "fallback")
	require.NoError(t, err)

	raw, err := fs.ReadFile(fallback, "index.html")
	require.NoError(t, err)
	assert.Contains(t, string(raw), "Periscope frontend assets are not built.")
}
