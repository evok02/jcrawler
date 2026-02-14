package main

import (
	"github.com/evok02/jcrawler/internal/config"
	"github.com/evok02/jcrawler/internal/db"
	"github.com/evok02/jcrawler/internal/parser"
	"github.com/evok02/jcrawler/internal/worker"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	godotenv.Load()
	cfg, err := config.NewConfig(".")
	if err != nil {
		log.Fatal(err.Error())
	}

	w := worker.NewWorker(cfg.Worker.Delay, cfg.Worker.Timeout)
	p := parser.NewParser(cfg.Keywords)

	s, err := db.NewStorage(cfg.DB.ConnString)
	if err != nil {
		log.Fatal(err.Error())
	}
	if err := s.Init(); err != nil {
		log.Fatal(err.Error())
	}
	defer s.CloseConnection()

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
