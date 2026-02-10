package parser

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"strings"
	"testing"
)

var keywords = []Keyword{
	Keyword{value: "Go", weight: 3},
	Keyword{value: "Intern", weight: 4},
	Keyword{value: "Internship", weight: 4},
	Keyword{value: "Backend", weight: 3},
}

func TestFindLinks(t *testing.T) {
	goodHtml := "<div class=\"section\"><ul><li><a href=\"url_found\"></li></ul></div>" //url_found
	emptyHtml := "<div class=\"section\"><ul><li><a href=\"\"></li></ul></div>"         //empty url
	parser := NewParser(keywords)

	//Test: GOOD HTML
	root, err := html.Parse(strings.NewReader(goodHtml))
	require.NoError(t, err)
	links := parser.findLinks(root)
	assert.Equal(t, 1, len(links))
	assert.Equal(t, links[0].url, "url_found")

	//Test: EMPTY HTML
	root, err = html.Parse(strings.NewReader(emptyHtml))
	require.NoError(t, err)
	links = parser.findLinks(root)
	assert.Equal(t, 1, len(links))
	assert.Equal(t, links[0].url, "")

}
