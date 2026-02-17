package filter

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/evok02/jcrawler/internal/db"
	"hash"
	"net/url"
	"strings"
	"sync"
	"time"
)

var ERROR_MALICIOUS_URL_FORMAT = errors.New("malicious url format")

type Filter struct {
	timeout time.Duration
	mu      sync.Mutex
	hash    hash.Hash
}

func NewFilter(t time.Duration) *Filter {
	return &Filter{
		timeout: t,
		hash:    sha256.New(),
	}
}

func (f *Filter) IsValid(link *url.URL, s *db.Storage) (bool, error) {
	parsedURLStr := link.String()

	if parsedURLStr == "" {
		return false, ERROR_MALICIOUS_URL_FORMAT
	}

	if strings.HasPrefix(parsedURLStr, "#") {
		return false, ERROR_MALICIOUS_URL_FORMAT
	}

	if strings.HasPrefix(parsedURLStr, "file:") {
		return false, ERROR_MALICIOUS_URL_FORMAT
	}

	if strings.HasPrefix(parsedURLStr, "javasript:") {
		return false, ERROR_MALICIOUS_URL_FORMAT
	}

	if strings.HasPrefix(parsedURLStr, "mailto:") {
		return false, ERROR_MALICIOUS_URL_FORMAT
	}

	hashed, err := f.HashLink(normalizeURL(parsedURLStr))
	if err != nil {
		return false, fmt.Errorf("checkTimeout: %s", err.Error())
	}

	return f.checkTimeout(s, string(hashed)), nil
}

func (f *Filter) HashLink(link string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.hash.Reset()
	_, err := f.hash.Write([]byte(link))
	if err != nil {
		return "", fmt.Errorf("hashLink: %s", err.Error())
	}

	return string(f.hash.Sum(nil)), nil
}

func (f *Filter) checkTimeout(s *db.Storage, id string) bool {
	p, err := s.GetPageByID(id)
	if err != nil {
		return true
	}

	if time.Since(p.UpdatedAt) < f.timeout {
		return false
	}

	return true
}

func normalizeURL(url string) string {
	normalized := strings.TrimSuffix(url, "/")
	return normalized
}
