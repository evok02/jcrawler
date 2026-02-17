package filter

import (
	"github.com/evok02/jcrawler/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
	"time"
)

var f = NewFilter(time.Hour * 6)

func TestIsValid(t *testing.T) {
	s, err := db.NewStorage("mongodb://localhost:27017/?retryWrites=true")
	require.NoError(t, err)
	defer s.CloseConnection()

	link, err := url.Parse("mailto://www.youtube.com/")
	require.NoError(t, err)
	b, err := f.IsValid(link, s)
	require.Error(t, err)
	assert.False(t, b)

	urlHash, err := f.HashLink("https://www.youtube.com")
	require.NoError(t, err)
	var p = db.Page{
		URLHash:       string(urlHash),
		Index:         0,
		KeywordsFound: []string{},
		UpdatedAt:     time.Now().UTC(),
	}

	require.NoError(t, s.InsertPage(&p))
	link, err = url.Parse("https://www.youtube.com/")
	require.NoError(t, err)
	b, err = f.IsValid(link, s)
	require.NoError(t, err)
}
