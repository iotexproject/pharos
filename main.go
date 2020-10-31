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
	"github.com/iotexproject/pharos/staking_v1"
)

func main() {
	flag.Parse()
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8192"
	}
	log.Println("======= starting pharos service")
	log.Println("======= port:", port)
	log.Println("======= TLS enabled:", handler.TLSEnabled())
	log.Println("======= blockchain endpoint:", handler.Endpoint())

	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
		return
	})
	r.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
		return
	})

	account := r.PathPrefix("/v1/accounts").Subrouter()
	account.HandleFunc("/{addr:[0-9ac-z]{41}}", handler.GrpcToHttpHandler(handler.GetAccount)).Methods(http.MethodGet)

	action := r.PathPrefix("/v1/actions").Subrouter()
	action.HandleFunc("/hash/{hash:[0-9a-fA-F]{64}}", handler.GrpcToHttpHandler(handler.GetActionByHash)).Methods(http.MethodGet)
	action.HandleFunc("/addr/{addr:[0-9ac-z]{41}}", handler.GrpcToHttpHandler(handler.GetActionByAddr)).Methods(http.MethodGet).
		Queries("start", "{start:[0-9]+}", "count", "{count:[0-9]+}")

	send := r.PathPrefix("/v1/actionbytes").Subrouter()
	send.HandleFunc("/{signedbytes:[0-9a-fA-F]+}", handler.GrpcToHttpHandler(handler.SendSignedActionBytes)).Methods(http.MethodPost)

	notification := r.PathPrefix("/v1/transfers").Subrouter()
	notification.HandleFunc("/block/{block:[0-9]+}", handler.GrpcToHttpHandler(handler.GetTsfInBlock)).Methods(http.MethodGet)

	meta := r.PathPrefix("/v1/chainmeta").Subrouter()
	meta.HandleFunc("", handler.GrpcToHttpHandler(handler.GetChainMeta)).Methods(http.MethodGet)

	staking := r.PathPrefix("/v1/staking").Subrouter()
	staking.HandleFunc("/validators", staking_v1.MemberValidators).Methods(http.MethodGet)
	staking.HandleFunc("/delegations/{addr:[0-9ac-z]{41}}", staking_v1.MemberDelegations).Methods(http.MethodGet)

	votes := r.PathPrefix("/v1/votes").Subrouter()
	votes.HandleFunc("/addr/{addr:[0-9ac-z]{41}}", handler.GrpcToHttpHandler(handler.GetVotesByAddr)).Methods(http.MethodGet)
	votes.HandleFunc("/index/{index:[0-9]+}", handler.GrpcToHttpHandler(handler.GetVotesByIndex)).Methods(http.MethodGet)

	contract := r.PathPrefix("/v1/contract").Subrouter()
	contract.HandleFunc("/addr/{addr:[0-9ac-z]{41}}", handler.GrpcToHttpHandler(handler.ReadContract)).Methods(http.MethodGet).
		Queries("method", "{method:[0-9a-fA-F]{8}}", "data", "{data:[0-9a-fA-F]+}")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 50 * time.Second,
		ReadTimeout:  50 * time.Second,
	}

	ln, err := httputil.LimitListener(srv.Addr)
	if err != nil {
		log.Fatal("======= error creating listener: ", err.Error())
	}
	log.Fatal(srv.Serve(ln))
}
