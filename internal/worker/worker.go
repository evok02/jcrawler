package worker

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type workerStatus int

const (
	Default workerStatus = iota
	Blocked
)

type Worker struct {
	delay            time.Duration
	timeout          time.Duration
	retriesCount     int
	maxRetriesAmount int
	status           workerStatus
}

func NewWorker(delay, timeout time.Duration) *Worker {
	return &Worker{
		delay:            delay,
		timeout:          timeout,
		retriesCount:     0,
		maxRetriesAmount: 3,
	}
}

func (w *Worker) Fetch(url string) (io.Reader, error) {
	req, err := createReqeust(url)
	if err != nil {
		return nil, fmt.Errorf("Test Job: %s", err)
	}

	res, err := sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("Test Job: %s", err)
	}

	//TODO: implement request after delay
	//if res.StatusCode >= 400 {
	//w.status = Blocked
	//for w.retriesCount < w.maxRetriesAmount && w.status == Blocked {
	//time.Sleep(w.delay)
	//}
	//}

	return res.Body, nil
}

func setHeaders(r *http.Request) {
	r.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	r.Header.Set("Accept-Language", "en-US,en;q=0.5")
	r.Header.Set("Accept-Encoding", "utf-8")
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Upgrade-Insecure-Requests", "1")
	r.Header.Set("Sec-Fetch-Dest", "document, navigate, same-origin")
	r.Header.Set("Cache-Control", "max-age=0")
	r.Header.Set("DNT", "1")
}

func createReqeust(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("createReqeust: %s", err)
	}
	setHeaders(req)
	return req, nil
}

func sendRequest(req *http.Request) (*http.Response, error) {
	res, err := new(http.Client).Do(req)
	if err != nil {
		return nil, fmt.Errorf("sendRequest: %s", err.Error())
	}
	return res, nil
}
