package parser

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"net/url"
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
	goodHtml := "<div class=\"section\"><ul><li><a href=\"www.youtube.com\"></li></ul></div>" //empty url
	emptyHtml := "<div class=\"section\"><ul><li><a href=\"\"></li></ul></div>"               //empty url
	sameHtml := "<div class=\"section\"><ul><li><a href=\"www.google.com\"></li></ul></div>"  //empty url
	parser := NewParser()
	url, err := url.Parse("www.google.com")
	require.NoError(t, err)
	parser.currAddr = url

	// Test: GOOD HTML
	root, err := html.Parse(strings.NewReader(goodHtml))
	require.NoError(t, err)
	parser.findLinks(root)
	assert.Equal(t, 1, len(parser.linksFound))
	assert.Equal(t, parser.linksFound[0].String(), "www.youtube.com")

	// Test: SAME HTML
	root, err = html.Parse(strings.NewReader(sameHtml))
	require.NoError(t, err)
	parser.findLinks(root)
	assert.Equal(t, 0, len(parser.linksFound))

	// Test: EMPTY HTML
	root, err = html.Parse(strings.NewReader(emptyHtml))
	require.NoError(t, err)
	parser.findLinks(root)
	assert.Equal(t, 0, len(parser.linksFound))
}

func TestKeywordsFound(t *testing.T) {
	noKeywordsHtml := "<div class=\"section\"><ul><li><a href=\"\"></li></ul></div>"
	nestedDivHtml := "<div><div>Go</div><div>Intern</div><div><p>Backend</p></div></div>"
	keywordAsClassNames := "<div><div class=\"Go\"></div><div>Intern</div><div><p class=\"Backend\"></p></div></div>"
	//p := NewParser(keywords)

	// Test: NoKeywords
	_, err := html.Parse(strings.NewReader(noKeywordsHtml))
	require.NoError(t, err)

	// Test: Nested Div
	_, err = html.Parse(strings.NewReader(nestedDivHtml))
	require.NoError(t, err)

	// Test: Keywords As Class Names
	_, err = html.Parse(strings.NewReader(keywordAsClassNames))
	require.NoError(t, err)
}
