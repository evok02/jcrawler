package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/evok02/jcrawler/internal/config"
	"github.com/evok02/jcrawler/internal/db"
	"github.com/evok02/jcrawler/internal/filter"
	"github.com/evok02/jcrawler/internal/index"
	"github.com/evok02/jcrawler/internal/parser"
	"github.com/evok02/jcrawler/internal/scheduler"
	"github.com/evok02/jcrawler/internal/worker"
	"github.com/joho/godotenv"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"
)

const MAX_AMOUNT_ROUTINES = 100

var ERROR_INVALID_URL_FORMAT = errors.New("malicious url format")

type App struct {
	Count    atomic.Int32
	ErrCount atomic.Int32
	Ctx      context.Context
	Worker   *worker.Worker
	Queue    *scheduler.JobQueue
	Filter   *filter.Filter
	Parser   *parser.Parser
	DB       *db.Storage
	Cfg      *config.Config
	Index    *index.Index
	Logger   *slog.Logger
}

func NewApp(cfgPath string) (*App, error) {
	app := &App{}
	godotenv.Load()
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	app.Cfg = cfg

	app.Worker = worker.NewWorker(cfg.Worker.Delay, cfg.Worker.Timeout)
	app.Parser = parser.NewParser()
	app.Filter = filter.NewFilter(time.Hour * 6)

	idx, err := index.Init(cfg.Index)
	if err != nil {
		return nil, err
	}
	app.Index = idx

	s, err := db.NewStorage(cfg.DB.ConnString)
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	app.DB = s
	app.Queue = scheduler.NewJobQueue(10)
	app.Ctx = context.Background()
	return app, nil
}

func (app *App) FetcherRoutine() <-chan *worker.FetchResponse {
	resChan := make(chan *worker.FetchResponse)
	sem := make(chan struct{}, MAX_AMOUNT_ROUTINES)
	go func() {
	outer:
		for {
			select {
			case url := <-app.Queue.Pop(app.Ctx, 1):
				sem <- struct{}{}
				go func() {
					defer func() { <-sem }()
					context, cancel := context.WithTimeout(context.Background(), time.Second*10)
					defer cancel()
					start := time.Now()
					res, err := app.Worker.Fetch(context, url)
					if err != nil {
						app.handleBadResponse(url, start, err)
						return
					}
					app.handleGoodResponse(url, start)
					resChan <- res
				}()
			case <-app.Ctx.Done():
				break outer
			}
		}
		close(resChan)
	}()
	return resChan
}

func (app *App) handleGoodResponse(url string, start time.Time) {
	app.Count.Add(1)
	app.Logger.Info("resource was fetched successfuly",
		slog.String("method", "GET"),
		slog.String("url", url),
		slog.Float64("response_time", time.Since(start).Seconds()))
}

func (app *App) handleBadResponse(url string, start time.Time, err error) {
	app.ErrCount.Add(1)
	app.Logger.Error("FetcherRoutine: %s"+err.Error(),
		slog.String("method", "GET"),
		slog.String("url", url),
		slog.Float64("response_time", time.Since(start).Seconds()))
}

func (app *App) parseResToPage(pres *parser.ParseResponse) (*db.Page, error) {
	if pres.Addr == nil {
		return nil, ERROR_INVALID_URL_FORMAT
	}
	hashLink, err := app.Filter.HashLink(pres.Addr.String())
	if err != nil {
		return nil, fmt.Errorf("ParseResToPage: %s", err.Error())
	}
	return &db.Page{
		URLHash:   hashLink,
		URL:       pres.Addr.String(),
		UpdatedAt: time.Now().UTC(),
		Content:   strings.ToValidUTF8(string(pres.Content), ""),
		Title:     pres.Title,
	}, nil
}

func (app *App) ParserRoutine(in <-chan *worker.FetchResponse) <-chan *parser.ParseResponse {
	resChan := make(chan *parser.ParseResponse)
	go func() {
	outer:
		for {
			select {
			case res := <-in:
				pres, err := app.Parser.Parse(res)
				if err != nil {
					go app.handleBadPage(res, err)
					continue outer
				}
				resChan <- pres
				go app.createEntry(pres)
			case <-app.Ctx.Done():
				break outer
			}
		}
		close(resChan)
	}()
	return resChan
}

func (app *App) handleBadPage(res *worker.FetchResponse, err error) {
	app.Logger.Error("ParserRoutine: %s"+err.Error(),
		slog.Any("res", res))
	app.ErrCount.Add(1)
}

func (app *App) createEntry(pres *parser.ParseResponse) {
	page, err := app.parseResToPage(pres)
	if err != nil {
		app.Logger.Error("ParserRoutine: %s"+err.Error(),
			slog.Any("page", page))
		app.ErrCount.Add(1)
		return
	}
	go func() {
		err = app.DB.InsertPage(page)
		if err != nil {
			slog.Error("ParserRoutine: %s"+err.Error(),
				slog.Any("page", page))
			app.ErrCount.Add(1)
			return
		}
	}()

	go func() {
		err := app.Index.HandleEntry(app.Ctx, page)
		if err != nil {
			app.Logger.Error("ParserRoutine: %s"+err.Error(),
				slog.Any("page", page))
			app.ErrCount.Add(1)
			return
		}
	}()
}

func (app *App) FilterRoutine(in <-chan *parser.ParseResponse) {
	go func() {
	outer:
		for {
			select {
			case res := <-in:
				linkChan := make(chan string)
				app.enqueIfValid(linkChan, res)
				close(linkChan)
			case <-app.Ctx.Done():
				break outer
			}
		}
	}()
}

func (app *App) enqueIfValid(linkChan chan string, res *parser.ParseResponse) {
	go func() {
		for range linkChan {
			app.Queue.Push(app.Ctx, 1, linkChan)
		}
	}()
	for _, link := range res.Links {
		if ok, err := app.Filter.IsValid(link, app.DB); ok && err == nil {
			linkChan <- link.String()
		}
	}
}

func (app *App) PushSeed() chan string {
	resChan := make(chan string)
	go func() {
		for _, link := range app.Cfg.Seed {
			resChan <- link
		}
		close(resChan)
	}()
	return resChan
}
