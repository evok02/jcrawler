package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/evok02/jcrawler/internal/config"
	"github.com/evok02/jcrawler/internal/db"
	"github.com/evok02/jcrawler/internal/filter"
	"github.com/evok02/jcrawler/internal/parser"
	"github.com/evok02/jcrawler/internal/scheduler"
	"github.com/evok02/jcrawler/internal/worker"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"os"
	"sync/atomic"
	"time"
)

var ERROR_INVALID_URL_FORMAT = errors.New("malicious url format")

const MAX_AMOUT_ROUTINES = 250

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
	logger   *slog.Logger
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
	sem := make(chan struct{}, MAX_AMOUT_ROUTINES)
	go func() {
	outer:
		for {
			select {
			case url := <-app.q.Pop(app.ctx, 1):
				sem <- struct{}{}
				go func() {
					defer func() { <-sem }()
					context, cancel := context.WithTimeout(context.Background(), time.Second*10)
					defer cancel()
					start := time.Now()
					res, err := app.w.Fetch(context, url)
					if err != nil {
						app.logger.Error("FetcherRoutine: %s"+err.Error(),
							slog.String("method", "GET"),
							slog.String("url", url),
							slog.Float64("response_time", time.Since(start).Seconds()))
						app.errCount.Add(1)
						return
					}
					app.count.Add(1)
					app.logger.Info("resource was fetched successfuly",
						slog.String("method", "GET"),
						slog.String("url", url),
						slog.Float64("response_time", time.Since(start).Seconds()))
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
	if pres.Addr == nil {
		return nil, ERROR_INVALID_URL_FORMAT
	}
	hashLink, err := app.f.HashLink(pres.Addr.String())
	if err != nil {
		return nil, fmt.Errorf("ParseResToPage: %s", err.Error())
	}
	keywords := []string{}
	for k, v := range pres.Matches.MatchesFound {
		if v == parser.FoundState {
			keywords = append(keywords, k)
		}
	}
	return &db.Page{
		URLHash:       hashLink,
		URL:           pres.Addr.String(),
		Index:         pres.Index,
		KeywordsFound: keywords,
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

func (app *App) ParserRoutine(in <-chan *worker.FetchResponse) <-chan *parser.ParseResponse {
	resChan := make(chan *parser.ParseResponse)
	pres := &parser.ParseResponse{}
	res := &worker.FetchResponse{}
	var err error
	go func() {
	outer:
		for {
			select {
			case res = <-in:
				pres, err = app.p.Parse(res)
				if err != nil {
					slog.Error("ParserRoutine: %s"+err.Error(),
						slog.Any("res", res))
					app.errCount.Add(1)
					continue outer
				}
				if pres.Index > 5 {
					log.Printf("Found valuable resource: %s\n", pres.Addr)
				}
				resChan <- pres
				go func() {
					page, err := app.ParseResToPage(pres)
					if err != nil {
						slog.Error("ParserRoutine: %s"+err.Error(),
							slog.Any("page", page))
						app.errCount.Add(1)
					}
					err = app.db.InsertPage(page)
					if err != nil {
						slog.Error("ParserRoutine: %s"+err.Error(),
							slog.Any("page", page))
						app.errCount.Add(1)
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
	f, err := os.OpenFile(app.cfg.Log.Path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()
	app.logger = slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{
		AddSource: true,
	}))

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
