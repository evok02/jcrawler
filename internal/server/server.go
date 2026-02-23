package server

import (
	"strings"
	"encoding/json"
	"net/http"
	"fmt"
	"github.com/evok02/jcrawler/internal/db"
	"reflect"
	"errors"
	"os"
)

const DEFAULT_ADDR string = "localhost:1337"

var ERROR_RES_NOT_POINTER =  errors.New("result should be a pointer")
var ERROR_MALFORMED_QUERY = errors.New("invalid query format")
var ERROR_EMPTY_CONN_STRING = errors.New("empty db connection string")

type ResponseError struct {
	Err error
}

func NewResponseError(err error) *ResponseError {
	return &ResponseError{
		Err: err,
	}
}

func (re *ResponseError) Error() string {
	return re.Err.Error()
}

func WriteJSON(w http.ResponseWriter, result any) error {
	resultVal := reflect.ValueOf(result)
	if resultVal.Kind() != reflect.Pointer {
		return fmt.Errorf("WriteJSON: %s", ERROR_RES_NOT_POINTER)
	}
	
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("WriteJSON: %s", err.Error())
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonResult)
	if err != nil {
		return fmt.Errorf("WriteJSON: %s", err.Error())
	}

	switch resultVal.Type(){
	case reflect.TypeOf(&ResponseError{}):
		w.WriteHeader(http.StatusBadRequest)
	case reflect.TypeOf([]db.PageServe{}):
		w.WriteHeader(http.StatusOK)
	}  

	return nil
}

type Server struct {
	srv http.Server
	db db.Storage
}

func New(addr string) *Server {
	return &Server{
		srv: http.Server {
			Addr: addr,
		},
	}	
}

type ApiConfig struct {
	store *db.Storage
}

func NewApiConfig() (*ApiConfig, error) {
	dbConn := os.Getenv("DB_CONN_STRING")
	if dbConn == "" {
		return nil, ERROR_EMPTY_CONN_STRING
	}
	store, err := db.NewStorage(dbConn)

	if err != nil {
		return nil, fmt.Errorf("NewApiConfig: %s", err.Error())
	}
	return &ApiConfig{
		store: store,
	}, nil
}

func (cfg *ApiConfig) HandleGetPages(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	if queries["search"] == nil {
		WriteJSON(w, NewResponseError(ERROR_MALFORMED_QUERY))
		return 
	}

	pages, err := cfg.store.GetPagesByIndex(strings.Join(queries["search"], " "))
	if err != nil {
		WriteJSON(w, NewResponseError(err))
		return 
	}

	WriteJSON(w, pages)
}

func Run(addr string) error {
	mux := http.NewServeMux()
	apiCfg, err := NewApiConfig()
	if err != nil {
		return fmt.Errorf("Run: %s", err.Error())
	}
	mux.HandleFunc("GET /api/page", apiCfg.HandleGetPages)
	//mux.HandleFunc("/api/limit")
	//mux.HandleFunc("/api/seed")

	server := New(addr)
	server.srv.Handler = mux
	return server.srv.ListenAndServe()
}
