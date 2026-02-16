package parser

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/http"
	"strings"
	"sync"
)

type matchState int

const (
	UninitializedState matchState = iota
	InitializedState
	FoundState
)

var ERROR_FOUND_TO_UNINIT_KEYWORD = errors.New("trying to set FoundState to unitialized value")

type Matches struct {
	MatchesFound map[string]matchState
}

func NewMatches() *Matches {
	return &Matches{
		MatchesFound: make(map[string]matchState),
	}
}

func (m *Matches) InitKeywords(k []string) {
	for _, value := range k {
		m.InitKeyword(value)
	}
}

func (m *Matches) Get(key string) (matchState, bool) {
	if v := m.MatchesFound[strings.ToLower(key)]; v == UninitializedState {
		return UninitializedState, false
	} else {
		return v, true
	}
}

func (m *Matches) SetFound(key string) error {
	if _, ok := m.Get(key); !ok {
		return ERROR_FOUND_TO_UNINIT_KEYWORD
	}
	m.MatchesFound[strings.ToLower(key)] = FoundState
	return nil
}

func (m *Matches) InitKeyword(key string) {
	m.MatchesFound[strings.ToLower(key)] = InitializedState
}

type Link struct {
	URL string
}

func NewLink(url string) Link {
	return Link{
		URL: url,
	}
}

type Parser struct {
	keywords   []string
	matches    *Matches
	linksFound []Link
}

func NewParser(keywords []string) *Parser {
	matches := NewMatches()
	matches.InitKeywords(keywords)
	return &Parser{
		keywords: keywords,
		matches:  matches,
	}
}

type ParseResponse struct {
	Links   []Link
	Addr    Link
	mu      *sync.Mutex
	Matches *Matches
	Index   int
}

var mu = new(sync.Mutex)

func (p *Parser) Parse(res *http.Response) (*ParseResponse, error) {
	pres := ParseResponse{mu: mu, Addr: NewLink(res.Request.URL.String())}
	root, err := html.Parse(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Parse: %s", err.Error())
	}

	p.findLinks(root)
	p.findMatches(root)
	pres.Index = calculateIndex(p.matches)
	pres.Links = p.linksFound
	pres.Matches = p.matches
	return &pres, nil
}

func (p *Parser) findLinks(root *html.Node) {
	p.linksFound = []Link{}
	for node := range root.Descendants() {
		if node.Type == html.ElementNode && node.DataAtom == atom.A {
			for _, a := range node.Attr {
				if a.Key == "href" {
					p.linksFound = append(p.linksFound, NewLink(a.Val))
				}
			}
		}
	}
}

func (p *Parser) findMatches(r *html.Node) error {
	p.matches = NewMatches()
	p.matches.InitKeywords(p.keywords)
	for node := range r.Descendants() {
		if node.Type == html.TextNode && node.DataAtom == 0 {
			for v := range strings.FieldsSeq(node.Data) {
				if state, ok := p.matches.Get(v); ok && state != FoundState {
					err := p.matches.SetFound(v)
					if err != nil {
						return fmt.Errorf("findMatches: %s", err.Error())
					}
				}
			}
		}
	}
	return nil
}

func calculateIndex(m *Matches) int {
	var res int
	for _, v := range m.MatchesFound {
		if v == FoundState {
			res++
		}
	}
	return res
}
