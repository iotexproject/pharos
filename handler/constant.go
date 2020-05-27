package handler

import (
	"crypto/tls"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	Mainnet = "api.iotex.one:443"
)

var (
	ErrGrpcConnFailed = "failed to establish gRPC connection"
	ErrGrpcCallFailed = "failed to execute gRPC call"

	endpoint  string
	enableTLS bool
)

func init() {
	endpoint = os.Getenv("IOTEX_ENDPOINT")
	if len(endpoint) == 0 {
		endpoint = Mainnet
	}

	tls := os.Getenv("ENABLE_TLS")
	enableTLS = (tls == "TRUE" || tls == "True" || endpoint == Mainnet)
}

func GrpcConnection() (*grpc.ClientConn, error) {
	if enableTLS {
		return grpc.Dial(endpoint, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}
	return grpc.Dial(endpoint, grpc.WithInsecure())
}
