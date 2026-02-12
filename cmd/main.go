package main

import (
	"github.com/evok02/jcrawler/internal/config"
	"github.com/evok02/jcrawler/internal/parser"
	"github.com/evok02/jcrawler/internal/storage"
	"github.com/evok02/jcrawler/internal/worker"
	"log"
)

func main() {
	cfg, err := config.NewConfig(".")
	if err != nil {
		log.Fatal(err.Error())
	}
	w := worker.NewWorker(cfg.Worker.Delay, cfg.Worker.Timeout)
	p := parser.NewParser(cfg.Keywords)
	s, err := storage.NewStorage(cfg.DB.ConnString)
	if err != nil {
		log.Fatal(err.Error())
	}
	if err := s.Init(); err != nil {
		log.Fatal(err.Error())
	}
	res, err := w.Fetch("https://at.indeed.com/jobs?q=Backend+Developer&l=%C3%96sterreich&ts=1770748402801&from=searchOnHP&rq=1&rsIdx=0&newcount=18&fromage=last&vjk=24294de2ddda237a")
	if err != nil {
		log.Fatal(err.Error())
	}
	stat, err := p.Parse(res)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Links: %+v\nAmount of matches: %+v\n", stat.Links, stat.Index)

}
