package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/iotexproject/go-pkgs/util/httputil"

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
	account.HandleFunc("/{addr:[0-9ac-z]{41}}", handler.GrpcToHttpHandler(handler.GetAccount)).Methods(http.MethodGet)

	action := r.PathPrefix("/v1/actions").Subrouter()
	action.HandleFunc("/hash/{hash:[0-9a-fA-F]{64}}", handler.GrpcToHttpHandler(handler.GetActionByHash)).Methods(http.MethodGet)
	action.HandleFunc("/addr/{addr:[0-9ac-z]{41}}", handler.GrpcToHttpHandler(handler.GetActionByAddr)).Methods(http.MethodGet).
		Queries("start", "{start:[0-9]+}", "count", "{count:[0-9]+}")

	send := r.PathPrefix("/v1").Subrouter()
	send.HandleFunc("/actionbytes/{signedbytes:[0-9a-fA-F]+}", handler.GrpcToHttpHandler(handler.SendSignedActionBytes)).Methods(http.MethodPost)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  25 * time.Second,
	}

	ln, err := httputil.LimitListener(srv.Addr)
	if err != nil {
		log.Fatal("======= error creating listener: ", err.Error())
	}
	log.Fatal(srv.Serve(ln))
}
