package main

import (
	"context"
	"github.com/evok02/jcrawler/internal/app"
	"log"
	"log/slog"
	"os"
	"time"
)

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

// TODO: add docker-compose file
// TODO: find a way to convert content to utf-8

// ---------------------------------------------------------

func main() {
	app, err := app.NewApp(".")
	if err != nil {
		log.Fatal(err.Error())
	}

	cancelContext, cancel := context.WithCancel(app.Ctx)
	defer cancel()
	defer app.DB.CloseConnection()
	f, err := os.OpenFile(app.Cfg.Log.Path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()
	app.Logger = slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{
		AddSource: true,
	}))

	app.Queue.Push(cancelContext, len(app.Cfg.Seed), app.PushSeed())
	httpResChan := app.FetcherRoutine()
	parseResChan := app.ParserRoutine(httpResChan)
	app.FilterRoutine(parseResChan)
	longTicker := time.NewTicker(time.Second * 100)
	shortTicker := time.NewTicker(time.Second * 10)
	go func() {
		count := 1
		for range longTicker.C {
			log.Printf("Total requests made in %d s: %d\nErrors made: %d\n",
				count*100, app.Count.Load(), app.ErrCount.Load())
			count++
		}
	}()
	go func() {
		for range shortTicker.C {
			log.Printf("Total request made: %d\nErrors made: %d\n",
				app.Count.Load(), app.ErrCount.Load())
		}
	}()
	log.Printf("Running...")
	time.Sleep(60 * time.Minute)
}
