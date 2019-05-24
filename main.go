package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/iotexproject/pharos/handler"
)

func main() {
	flag.Parse()
	log.Println("======= starting pharos service")

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8192"
	}

	r := mux.NewRouter()

	account := r.PathPrefix("/v1/accounts").Subrouter()
	account.HandleFunc("/{addr}", handler.GrpcToHttpHandler(handler.GetAccount)).Methods(http.MethodGet)

	action := r.PathPrefix("/v1/actions").Subrouter()
	action.HandleFunc("/hash/{hash}", handler.GrpcToHttpHandler(handler.GetActionByHash)).Methods(http.MethodGet)
	action.HandleFunc("/addr/{addr}", handler.GrpcToHttpHandler(handler.GetActionByAddr)).Methods(http.MethodGet).
		Queries("start", "{start:[0-9]+}", "count", "{count:[0-9]+}")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  25 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
