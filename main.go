package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"os"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
)

const (
	Mainnet = "api.iotex.one:443"
)

func run() error {
	endpoint := os.Getenv("IOTEX_ENDPOINT")
	if len(endpoint) == 0 {
		endpoint = Mainnet
	}
	glog.Warningln("======= endpoint: ", endpoint)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8192"
	}

	var dialOption grpc.DialOption
	enableTLS := os.Getenv("TLS_ENABLED")
	if enableTLS == "TRUE" || enableTLS == "True" || endpoint == Mainnet {
		dialOption = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
		glog.Warningln("======= TLS enabled")
	} else {
		dialOption = grpc.WithInsecure()
		glog.Warningln("======= TLS not enabled")
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{dialOption}
	err := iotexapi.RegisterAPIServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(":"+port, mux)
}

func main() {
	flag.Parse()
	glog.Infoln("======= starting pharos service")
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
