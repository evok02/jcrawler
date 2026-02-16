package filter

import (
	"github.com/evok02/jcrawler/internal/db"
	"github.com/evok02/jcrawler/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var f = NewFilter(time.Hour * 6)

func TestIsUrl(t *testing.T) {
	s, err := db.NewStorage("mongodb://localhost:27017/?retryWrites=true")
	require.NoError(t, err)
	defer s.CloseConnection()

	b := f.isUrl("https://www.youtube.com/")
	assert.True(t, b)

	b = f.isUrl("www.youtube.com/1?number=2")
	assert.True(t, b)

	b = f.isUrl("youtube.com/")
	assert.True(t, b)

	b = f.isUrl("/pages/1?number=2")
	assert.False(t, b)

	b = f.isUrl("select * from db")
	assert.False(t, b)
}

func TestIsValid(t *testing.T) {
	s, err := db.NewStorage("mongodb://localhost:27017/?retryWrites=true")
	require.NoError(t, err)
	defer s.CloseConnection()

	b, err := f.IsValid(parser.NewLink("https://www.youtube.com/"), s)
	require.NoError(t, err)
	assert.True(t, b)

	urlHash, err := f.HashLink("https://www.youtube.com")
	require.NoError(t, err)
	var p = db.Page{
		URLHash:       string(urlHash),
		Index:         0,
		KeywordsFound: []string{},
		UpdatedAt:     time.Now().UTC(),
	}

	require.NoError(t, s.InsertPage(&p))
	b, err = f.IsValid(parser.NewLink("https://www.youtube.com"), s)
	require.NoError(t, err)
	assert.False(t, b)

	//n, err := time.Parse("2006-Jan-02", "2026-Feb-14")
	//require.NoError(t, err)
	//urlHash, err = f.hashLink("https://www.google.com")
	//require.NoError(t, err)
	//
	//var np = db.Page{
	//URLHash:       string(urlHash),
	//Index:         0,
	//KeywordsFound: []string{},
	//UpdatedAt:     n.UTC(),
	//}
	//
	//require.NoError(t, s.InsertPage(&np))
	//s.GetPageByID(string(urlHash))
	//b, err = f.IsValid(parser.NewLink("https://www.google.com"), s)
	//require.NoError(t, err)
	//assert.True(t, b)
}
