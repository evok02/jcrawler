package filter

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/evok02/jcrawler/internal/db"
	"github.com/evok02/jcrawler/internal/parser"
	"hash"
	"regexp"
	"time"
)

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

	hashed, err := f.HashLink(url)
	if err != nil {
		return false, fmt.Errorf("checkTimeout: %s", err.Error())
	}

	return f.checkTimeout(s, string(hashed)), nil
}

func (f *Filter) isUrl(link string) bool {
	validUrl := regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/|\/|\/\/)?[A-z0-9_-]*?[:]?[A-z0-9_-]*?[@]?[A-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
	return validUrl.MatchString(link)
}

func (f *Filter) HashLink(link string) ([]byte, error) {
	f.hash.Reset()

	_, err := f.hash.Write([]byte(link))
	if err != nil {
		return nil, fmt.Errorf("hashLink: %s", err.Error())
	}

	return f.hash.Sum(nil), nil
}

func (f *Filter) checkTimeout(s *db.Storage, id string) bool {
	p, err := s.GetPageByID(id)
	if err != nil {
		return true
	}

	fmt.Printf("Id: %+v\nDate: %+v\n", id, p.UpdatedAt)
	if time.Since(p.UpdatedAt) < f.timeout {
		return false
	}

	return true
}
