package apiserver

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/UshakovN/practice/internal/app/parser"
	"github.com/UshakovN/practice/internal/app/store"
	"github.com/gorilla/mux"
)

type APIServer struct {
	config *Config
	router *mux.Router
}

func NewServer(config *Config) *APIServer {
	return &APIServer{
		config: config,
		router: mux.NewRouter(),
	}
}

func (server *APIServer) Start() error {
	log.Println("Server start")
	server.configureRouter()
	return http.ListenAndServe(server.config.BindAddr, server.router)
}

func (server *APIServer) configureRouter() {
	server.router.HandleFunc("/parsebrand", server.handleParseBrand()).Methods(http.MethodGet)
	server.router.HandleFunc("/", server.handleStart())
}

func (server *APIServer) handleStart() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("start"))
	}
}

func (server *APIServer) handleParseBrand() http.HandlerFunc {
	storeClient := store.NewClient(store.NewConfig())

	return func(rw http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		name, code := query.Get("name"), query.Get("code")
		if name == "" || code == "" {
			http.Error(rw, errors.New("invalid parameters").Error(), http.StatusBadRequest)
			return
		}
		brandParser := parser.NewParser(parser.Brand{
			Name: name,
			Code: code,
		})
		go brandParser.FisherSciencific(storeClient)

		resp := make(map[string]string)
		resp["message"] = http.StatusText(http.StatusOK)
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		rw.Write(jsonResp)
	}
}
