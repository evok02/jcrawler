package parser

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"strings"
	"testing"
)

var keywords = []string{
	"Go",
	"Intern",
	"Internship",
	"Backend",
	"Sofware Enginner",
}

func TestFindLinks(t *testing.T) {
	goodHtml := "<div class=\"section\"><ul><li><a href=\"url_found\"></li></ul></div>" //url_found
	emptyHtml := "<div class=\"section\"><ul><li><a href=\"\"></li></ul></div>"         //empty url
	parser := NewParser(keywords)

	// Test: GOOD HTML
	root, err := html.Parse(strings.NewReader(goodHtml))
	require.NoError(t, err)
	parser.findLinks(root)
	assert.Equal(t, 1, len(parser.linksFound))
	assert.Equal(t, parser.linksFound[0].URL, "url_found")

	// Test: EMPTY HTML
	root, err = html.Parse(strings.NewReader(emptyHtml))
	require.NoError(t, err)
	parser.findLinks(root)
	assert.Equal(t, 1, len(parser.linksFound))
	assert.Equal(t, parser.linksFound[0].URL, "")

}

func TestKeywordsFound(t *testing.T) {
	noKeywordsHtml := "<div class=\"section\"><ul><li><a href=\"\"></li></ul></div>"
	nestedDivHtml := "<div><div>Go</div><div>Intern</div><div><p>Backend</p></div></div>"
	keywordAsClassNames := "<div><div class=\"Go\"></div><div>Intern</div><div><p class=\"Backend\"></p></div></div>"
	p := NewParser(keywords)

	// Test: NoKeywords
	root, err := html.Parse(strings.NewReader(noKeywordsHtml))
	require.NoError(t, err)
	p.findMatches(root)
	for _, keyword := range keywords {
		v, ok := p.matches.Get(keyword)
		require.True(t, ok)
		require.NotEqual(t, FoundState, v)
		assert.Equal(t, InitializedState, v)
	}

	// Test: Nested Div
	root, err = html.Parse(strings.NewReader(nestedDivHtml))
	require.NoError(t, err)
	err = p.findMatches(root)
	require.NoError(t, err)
	v, ok := p.matches.Get("go")
	require.True(t, ok)
	require.Equal(t, FoundState, v)
	v, ok = p.matches.Get("intern")
	require.True(t, ok)
	require.Equal(t, FoundState, v)
	v, ok = p.matches.Get("backend")
	require.True(t, ok)
	require.Equal(t, FoundState, v)
	v, ok = p.matches.Get("python")
	require.False(t, ok)
	require.Equal(t, UninitializedState, v)

	// Test: Keywords As Class Names
	root, err = html.Parse(strings.NewReader(keywordAsClassNames))
	require.NoError(t, err)
	err = p.findMatches(root)
	require.NoError(t, err)
	v, ok = p.matches.Get("go")
	require.True(t, ok)
	require.Equal(t, v, InitializedState)
	v, ok = p.matches.Get("backend")
	require.True(t, ok)
	require.Equal(t, v, InitializedState)
	v, ok = p.matches.Get("intern")
	require.True(t, ok)
	require.Equal(t, v, FoundState)
}
