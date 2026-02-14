package filter

import (
	"crypto/sha256"
	"github.com/evok02/jcrawler/internal/parser"
	"hash"
	"regexp"
	"time"
)

// filter should check for link entry in the db
// filter should check if the provided entry is a link

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

func (f *Filter) isUrl(link string) bool {
	validUrl := regexp.MustCompile(`b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`)
	return validUrl.MatchString(link)
}

func (f *Filter) checkIfAlreadyExist(l parser.Link) (bool, error) {
	return false, nil
}
