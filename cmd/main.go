package main

import (
	"context"
	"github.com/evok02/jcrawler/internal/config"
	"github.com/evok02/jcrawler/internal/db"
	"github.com/evok02/jcrawler/internal/filter"
	"github.com/evok02/jcrawler/internal/parser"
	"github.com/evok02/jcrawler/internal/scheduler"
	"github.com/evok02/jcrawler/internal/worker"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"time"
	"sync/atomic"
)

var ctx = context.Background()

type App struct {
	count atomic.Int32
	ctx context.Context
	w   *worker.Worker
	q   *scheduler.JobQueue
	f   *filter.Filter
	p   *parser.Parser
	db  *db.Storage
	cfg *config.Config
}

func NewApp(cfgPath string) (*App, error) {
	app := &App{}
	godotenv.Load()
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	app.cfg = cfg

	app.w = worker.NewWorker(cfg.Worker.Delay, cfg.Worker.Timeout)
	app.p = parser.NewParser(cfg.Keywords)
	app.f = filter.NewFilter(time.Hour * 6)

	s, err := db.NewStorage(cfg.DB.ConnString)
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	app.db = s
	app.q = scheduler.NewJobQueue(1)
	app.ctx = context.Background()
	return app, nil
}

func (app *App) FetcherRoutine() <-chan *http.Response {
	resChan := make(chan *http.Response)
	go func() {
	outer:
		for {
			select {
			case url := <-app.q.Pop(ctx, 1):
				go func() {
					res, err := app.w.Fetch(url)
					if err != nil {
						log.Printf("FetcherRoutine: %s", err.Error())
						return
					}
					app.count.Add(1)
					resChan <- res
				}()
			case <-app.ctx.Done():
				break outer
			}
		}
		close(resChan)
	}()
	return resChan
}

func (app *App) ParseResToPage(pres *parser.ParseResponse) *db.Page {
	hashLink, err := app.f.HashLink(pres.Addr.URL)
	if err != nil {
		log.Printf("ParserRoutine: %s", err.Error())
	}
	keywords := []string{}
	for key := range pres.Matches.MatchesFound {
		keywords = append(keywords, key)
	}
	return &db.Page{
		URLHash:       string(hashLink),
		Index:         pres.Index,
		KeywordsFound: keywords,
		UpdatedAt:     time.Now().UTC(),
	}
}

func (app *App) ParserRoutine(in <-chan *http.Response) <-chan *parser.ParseResponse {
	resChan := make(chan *parser.ParseResponse)
	go func() {
	outer:
		for {
			select {
			case res := <-in:
				pres, err := app.p.Parse(res)
				if err != nil {
					log.Printf("ParserRoutine: %s", err.Error())
					continue outer
				}
				log.Printf("Parsed %s: %+v\n", pres.Addr.URL, pres)
				resChan <- pres
				go func() {
					page := app.ParseResToPage(pres)
					if err := app.db.InsertPage(page); err != nil {
						log.Printf("ParserRoutine: %s", err.Error())
					}
				}()
			case <-app.ctx.Done():
				break outer
			}
		}
		close(resChan)
	}()
	return resChan
}

func (app *App) FilterRoutine(in <-chan *parser.ParseResponse) {
	go func() {
	outer:
		for {
			select {
			case res := <-in:
				linkCh := make(chan string)
				go func() {
					for range linkCh {
						app.q.Push(app.ctx, 1, linkCh)
					}
				}()
				for _, link := range res.Links {
					if ok, err := app.f.IsValid(link, app.db); ok && err == nil {
						linkCh <- link.URL
					}
				}
				close(linkCh)
			case <-app.ctx.Done():
				break outer
			}
		}
	}()
}

func (app *App) PushSeed() chan string {
	resChan := make(chan string)
	go func() {
		for _, link := range app.cfg.Seed {
			resChan <- link
		}
		close(resChan)
	}()
	return resChan
}

func main() {
	app, err := NewApp(".")
	if err != nil {
		log.Fatal(err.Error())
	}

	cancelContext, cancel := context.WithCancel(app.ctx)
	defer cancel()
	defer app.db.CloseConnection()

	app.q.Push(cancelContext, len(app.cfg.Seed), app.PushSeed())
	httpResChan := app.FetcherRoutine()
	parseResChan := app.ParserRoutine(httpResChan)
	app.FilterRoutine(parseResChan)
	ticker := time.NewTicker(time.Second * 100)
	for range ticker.C{
		log.Printf("Total requests made in 100s: %d\n", app.count.Load())
	}
	time.Sleep(600 * time.Second)
}
