package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

type workerStatus int

var ERROR_RESPONSE_TIME_OVER_LIMIT = errors.New("response time is over the limit")
var ERROR_RETRIES_OVER_LIMIT = errors.New("worker reached maximum amount of returies")

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
	logger           *slog.Logger
}

type FetchResponse struct {
	Response *http.Response
	HostName *url.URL
}

func (fr *FetchResponse) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("url", fr.HostName.String()),
		slog.Group("response",
			slog.String("status", fr.Response.Status)),
		slog.Int64("content_length", fr.Response.ContentLength),
	)
}

func NewWorker(delay, timeout time.Duration) *Worker {
	return &Worker{
		delay:            delay,
		timeout:          timeout,
		retriesCount:     0,
		maxRetriesAmount: 3,
		logger:           slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (w *Worker) Fetch(ctx context.Context, url string) (*FetchResponse, error) {
	req, err := w.createReqeust(url)
	if err != nil {
		return nil, fmt.Errorf("Fetch %s", err)
	}
	resChan, errChan := w.sendRequest(req)
	for {
		select {
		case <-ctx.Done():
			return nil, ERROR_RESPONSE_TIME_OVER_LIMIT

		case err := <-errChan:
			return nil, err

		case res := <-resChan:
			return &FetchResponse{
				HostName: req.URL,
				Response: res,
			}, nil
		}
	}
}

func setHeaders(r *http.Request) {
	r.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")
	r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	r.Header.Set("Accept-Language", "en-US,en;q=0.5")
	r.Header.Set("Accept-Encoding", "utf-8")
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Upgrade-Insecure-Requests", "1")
	r.Header.Set("Sec-Fetch-Dest", "document, navigate, same-origin")
	r.Header.Set("Cache-Control", "max-age=0")
	r.Header.Set("DNT", "1")
}

func (w *Worker) createReqeust(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("createReqeust: %s", err)
	}
	setHeaders(req)
	return req, nil
}

func (w *Worker) sendRequest(req *http.Request) (chan *http.Response, chan error) {
	resChan := make(chan *http.Response)
	errChan := make(chan error)

	go func() {
		res, err := new(http.Client).Do(req)
		if err != nil {
			errChan <- fmt.Errorf("sendRequest: %s", err.Error())
		}
		resChan <- res
	}()

	return resChan, errChan
}
