package main

import (
	"context"
	"errors"
	"github.com/evok02/jcrawler/internal/config"
	"github.com/evok02/jcrawler/internal/db"
	"github.com/evok02/jcrawler/internal/filter"
	"github.com/evok02/jcrawler/internal/parser"
	"github.com/evok02/jcrawler/internal/scheduler"
	"github.com/evok02/jcrawler/internal/worker"
	"github.com/joho/godotenv"
	"log"
	"sync/atomic"
	"time"
)

// TODO: fix data integrity
// TODO: improve filtering to decrease error rate
var ERROR_INVALID_URL_FORMAT = errors.New("malicious url format")

type App struct {
	count    atomic.Int32
	errCount atomic.Int32
	ctx      context.Context
	w        *worker.Worker
	q        *scheduler.JobQueue
	f        *filter.Filter
	p        *parser.Parser
	db       *db.Storage
	cfg      *config.Config
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
	app.q = scheduler.NewJobQueue(10)
	app.ctx = context.Background()
	return app, nil
}

func (app *App) FetcherRoutine() <-chan *worker.FetchResponse {
	resChan := make(chan *worker.FetchResponse)
	go func() {
	outer:
		for {
			select {
			case url := <-app.q.Pop(app.ctx, 1):
				go func() {
					res, err := app.w.Fetch(url)
					if err != nil {
						log.Printf("FetcherRoutine: %s", err.Error())
						app.errCount.Add(1)
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

func (app *App) ParseResToPage(pres *parser.ParseResponse) (*db.Page, error) {
	if pres == nil {
		return nil, ERROR_INVALID_URL_FORMAT
	}
	hashLink, err := app.f.HashLink(pres.Addr.String())
	if err != nil {
		log.Printf("ParserRoutine: %s", err.Error())
	}
	keywords := []string{}
	for key := range pres.Matches.MatchesFound {
		keywords = append(keywords, key)
	}
	return &db.Page{
		URLHash:       hashLink,
		Index:         pres.Index,
		KeywordsFound: keywords,
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

func (app *App) ParserRoutine(in <-chan *worker.FetchResponse) <-chan *parser.ParseResponse {
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
				if pres.Index > 2 {
					log.Printf("Found valuable resource: %s\n", pres.Addr)
				}
				resChan <- pres
				go func() {
					page, err := app.ParseResToPage(pres)
					if err != nil {
						log.Printf("ParserRoutine: %s", err.Error())
					}
					err = app.db.InsertPage(page)
					if err != nil {
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
						linkCh <- link.String()
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
	longTicker := time.NewTicker(time.Second * 100)
	shortTicker := time.NewTicker(time.Second * 10)
	go func() {
		for range longTicker.C {
			log.Printf("Total requests made in 100s: %d\nFrom which errors: %d\n", app.count.Load(), app.errCount.Load())
		}
	}()
	go func() {
		for range shortTicker.C {
			log.Printf("Total request made: %d\nFrom which errors: %d\n", app.count.Load(), app.errCount.Load())
		}
	}()

	log.Printf("Running...")
	time.Sleep(100 * time.Second)
}
