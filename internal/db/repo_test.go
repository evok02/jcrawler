package db

import (
	"github.com/evok02/jcrawler/internal/config"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var p = Page{
	URLHash:       "google.com",
	Index:         2,
	KeywordsFound: []string{"google", "com"},
	UpdatedAt:     time.Now().UTC(),
}

func TestInsertGetPage(t *testing.T) {
	err := godotenv.Load("./../../.env")
	require.NoError(t, err)

	cfg, err := config.NewConfig("./../../")
	require.NoError(t, err)

	s, err := NewStorage(cfg.DB.ConnString)
	require.NoError(t, err)
	defer s.CloseConnection()

	err = s.Init()
	require.NoError(t, err)

	err = s.DeletePageByID("google.com")
	require.NoError(t, err)

	err = s.InsertPage(&p)
	require.NoError(t, err)

	page, err := s.GetPageByID("google.com")
	require.NoError(t, err)

	assert.Equal(t, p.URLHash, page.URLHash)
	assert.Equal(t, p.KeywordsFound, page.KeywordsFound)
	assert.Equal(t, p.Index, page.Index)
}

func TestDeletePage(t *testing.T) {
	err := godotenv.Load("./../../.env")
	require.NoError(t, err)

	cfg, err := config.NewConfig("./../../")
	require.NoError(t, err)

	s, err := NewStorage(cfg.DB.ConnString)
	require.NoError(t, err)
	defer s.CloseConnection()

	err = s.Init()
	require.NoError(t, err)

	err = s.DeletePageByID("google.com")
	require.NoError(t, err)
}

func TestUpdatePage(t *testing.T) {
	err := godotenv.Load("./../../.env")
	require.NoError(t, err)

	cfg, err := config.NewConfig("./../../")
	require.NoError(t, err)

	s, err := NewStorage(cfg.DB.ConnString)
	require.NoError(t, err)
	defer s.CloseConnection()

	err = s.Init()
	require.NoError(t, err)

	err = s.InsertPage(&p)
	require.NoError(t, err)

	res, err := s.UpdatePageByID("google.com", &Page{
		URLHash:       "google.com",
		KeywordsFound: []string{"google", "com", "ua"},
		Index:         3,
		UpdatedAt:     time.Now().UTC(),
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"google", "com", "ua"}, res.KeywordsFound)
	assert.Equal(t, 3, res.Index)
}
