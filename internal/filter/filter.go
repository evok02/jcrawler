package filter

import (
	"fmt"
	"crypto/sha256"
	"github.com/evok02/jcrawler/internal/parser"
	"hash"
	"regexp"
	"time"
	"github.com/evok02/jcrawler/internal/db"
	"errors"
)

// filter should check for link entry in the db
// filter should check if the provided entry is a link

var ERROR_MALICIOUS_URL_FORMAT = errors.New("malicious url format")

type Filter struct {
	timeout time.Duration
	hash    hash.Hash
}

func NewFilter(t time.Duration) *Filter {
	return &Filter{
		timeout: t,
		hash:    sha256.New(),
	}
}

func (f *Filter) IsValid(link parser.Link, s *db.Storage) (bool, error) {
	url := link.URL
	if !f.isUrl(url) {
		return false, ERROR_MALICIOUS_URL_FORMAT
	}

	return f.checkTimeout(s, url) 
}

func (f *Filter) isUrl(link string) bool {
	validUrl := regexp.MustCompile(`\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`)
	return validUrl.MatchString(link)
}

func (f *Filter) hashLink(link string) ([]byte, error) {
	_, err := f.hash.Write([]byte(link))
	if err != nil {
		return nil, fmt.Errorf("hashLink: %s", err.Error())
	}

	return f.hash.Sum(nil), nil
}

func (f *Filter) checkTimeout(s *db.Storage, link string) (bool, error) {
	hashed, err := f.hashLink(link)
	if err != nil {
		return false, fmt.Errorf("checkIfAlreadyExist: %s", err.Error())
	}

	p, err := s.GetPageByID(string(hashed))
	if err != nil {
		return false, fmt.Errorf("checkIfAlreadyExist: %s", err.Error())
	}

	if (time.Since(p.UpdatedAt) < f.timeout) {
		return false, nil
	}
	
	return true, nil
}
