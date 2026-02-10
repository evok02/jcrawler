package parser

import (
	"strings"
	"sync"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/http"
)

//TODO: create a data structure for matches

//TODO: create a data structure for keywords

//TODO: finish parse method

//TODO: finish find keywords method

//TODO: write calculateIndex method

type matches struct {
	matchesFound map[Keyword]int
}


func (m *matches) Set(key Keyword, value int) {
	if v := m.Get(key); v  == -1 {
		v = 1
	}
	m.matchesFound[key] = value
}

func (m *matches) Get(key Keyword) (int, bool) {
	return m.matchesFound[key]	
}

type Keywords struct {
	mu sync.Mutex
	values []Keyword
}

func (k *Keywords) Add(key Keyword) {
	mu.Lock()
	k.values = append(k.values, key)
	mu.Unlock()
}

func (m *matches) Init(k *Keywords) {
	for _, v := range k.values {
		m.Set(v, -1)
	}
}


type Link struct {
	url string
}

func newLink(url string) Link {
	return Link{
		url: url,
	}
}

type Keyword struct {
	value  string
	weight int
}

type Parser struct {
	keywords *Keywords
	response *ParseResponse
}

func NewParser(keywords *Keywords) *Parser {
	return &Parser{
		keywords: keywords,
	}
}

type ParseResponse struct {
	mu *sync.Mutex
	Index int
	Links []Link
}

func (pr *ParseResponse) updateIndex(idx int) {
	pr.Index = idx
}

func (pr *ParseResponse) appendLink(l Link) {
	pr.mu.Lock()
	pr.Links = append(pr.Links, l)
	pr.mu.Unlock()
}

var mu = new(sync.Mutex)

func (p *Parser) Parse(req *http.Request) (*ParseResponse, error) {
	res := ParseResponse{mu: mu,}
	p.response = &res
	root, err := html.Parse(req.Body)
	defer req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Parse: %s", err.Error())
	}
	res.Links = p.findLinks(root)
	return &res, nil
}

func (p *Parser) findLinks(root *html.Node) []Link {
	var links []Link
	for node := range root.Descendants() {
		if node.Type == html.ElementNode && node.DataAtom == atom.A {
			for _, a := range node.Attr {
				if a.Key == "href" {
					p.response.appendLink(newLink(a.Val))
				}
			}
		}
	}
	return links
}

func (p *Parser) findMatches(r *html.Node) matches {
	ms := matches{}
	ms.Init(p.keywords)
	for node := range r.Descendants() {
		if node.Type == html.TextNode && node.DataAtom == 0 {
			for _, v := range strings.Fields(node.Data) {
			}
		}
	}
	return nil
}

func calculateIndex(matches) int {
	return 0
}
