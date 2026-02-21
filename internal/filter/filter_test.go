package filter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var f = NewFilter(time.Hour * 6)

func TestNormalizeURL(t *testing.T) {
	normalized := normalizeURL("www.google.com//")
	assert.Equal(t, "www.google.com/", normalized)

	normalized = normalizeURL("/")
	assert.NotEqual(t, "", normalized)
}

func TestHashLink(t *testing.T) {
	hash, err := f.HashLink("www.google.com/")
	require.NoError(t, err)
	assert.NotEqual(t, "", hash)

	hashClone, err := f.HashLink("www.google.com/")
	require.NoError(t, err)
	assert.Equal(t, hash, hashClone)
}

func TestFilterLink(t *testing.T) {
	ok, err := filterLink("/")
	require.Error(t, err)
	assert.False(t, ok)

	ok, err = filterLink("#google.com")
	require.Error(t, err)
	assert.False(t, ok)

	ok, err = filterLink("file:howtobecomrich.pdf")
	require.Error(t, err)
	assert.False(t, ok)

	ok, err = filterLink("javascript:something.js")
	require.Error(t, err)
	assert.False(t, ok)

	ok, err = filterLink("mailto:1234@gmail.com")
	require.Error(t, err)
	assert.False(t, ok)

	ok, err = filterLink("https://netflix.com/")
	require.NoError(t, err)
	assert.True(t, ok)
}
