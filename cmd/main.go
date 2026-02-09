package main

import (
	"fmt"
	"gihtub.com/evok02/jcrawler/internal/worker"
	"io"
	"log"
	"time"
)

func main() {
	w := worker.NewWorker(3*time.Second, time.Second)
	body, err := w.Fetch("https://devjobs.at/job/cbbd17942a2b425fe70022bf19c475d3")
	if err != nil {
		fmt.Println(err.Error())
	}
	buf := make([]byte, 1024)
	for {
		n, err := body.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err.Error())
		}
		fmt.Printf("%s", buf)
		copy(buf, buf[n:])
	}
}
