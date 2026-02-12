package db

import (
	"github.com/evok02/jcrawler/internal/config"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"testing"
)

// Test: Connect to MongoDB instance

func TestConnect(t *testing.T) {
	err := godotenv.Load("./../../.env")
	require.NoError(t, err)

	cfg, err := config.NewConfig("./../../")
	require.NoError(t, err)

	s, err := NewStorage(cfg.DB.ConnString)
	require.NoError(t, err)

	err = s.Init()
	require.NoError(t, err)
}
