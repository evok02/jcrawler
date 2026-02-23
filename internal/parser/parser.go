package parser

import (
	"errors"
	"fmt"
	"github.com/evok02/jcrawler/internal/worker"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/url"
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

type Parser struct {
	buf        []byte
	currTitle  string
	linksFound []*url.URL
	currAddr   *url.URL
}

func NewParser() *Parser {
	return &Parser{
		buf: make([]byte, 4096),
	}
}

type ParseResponse struct {
	Content []byte
	Links   []*url.URL
	Title   string
	Addr    *url.URL
	mu      *sync.Mutex
}

var mu = new(sync.Mutex)

func (p *Parser) Parse(fres *worker.FetchResponse) (*ParseResponse, error) {
	pres := ParseResponse{mu: mu, Addr: fres.HostName}
	p.currAddr = fres.HostName
	if title := fres.Response.Header["Title"]; len(title) > 0 {
		pres.Title = title[0]
	}
	root, err := html.Parse(fres.Response.Body)
	if err != nil {
		return nil, fmt.Errorf("Parse: %s", err.Error())
	}
	defer fres.Response.Body.Close()

	p.findLinks(root)
	p.findRawText(root)
	pres.Links = p.linksFound
	pres.Content = append(pres.Content, p.buf...)
	pres.Title = p.currTitle
	p.currTitle = ""
	p.buf = p.buf[len(p.buf):]
	return &pres, nil
}

func (p *Parser) findLinks(root *html.Node) {
	p.linksFound = []*url.URL{}
	for node := range root.Descendants() {
		if node.Type == html.ElementNode && node.DataAtom == atom.A {
			for _, a := range node.Attr {
				if a.Key == "href" && len(a.Val) > 0 {
					parsed, err := url.Parse(a.Val)
					if err != nil {
						continue
					}
					if a.Val == p.currAddr.String() {
						continue
					}
					if a.Val[0] == '/' && len(a.Val) > 1 {
						p.linksFound = append(p.linksFound, parsed.ResolveReference(p.currAddr))
						continue
					}
					p.linksFound = append(p.linksFound, parsed)
				}
			}
		}
	}
}

func (p *Parser) findRawText(n *html.Node) {
	var traverse func(n *html.Node)
	traverse = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			normalized := strings.TrimSpace(n.Data)
			if len(normalized) != 0 {
				p.buf = append(p.buf, (normalized + " ")...)
			}
		case html.ElementNode:
			if n.Data == "script" || n.Data == "style" {
				return
			}
			if n.Data == "title" && n.FirstChild != nil {
				p.currTitle = n.FirstChild.Data
			}
		}
		for e := n.FirstChild; e != nil; e = e.NextSibling {
			traverse(e)
		}
	}
	traverse(n)
}
