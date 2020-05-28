package handler

import (
	"context"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
)

// GetAccount extracts address from http request, make gRPC call GetAccount()
func GetAccount(r *http.Request, cli iotexapi.APIServiceClient) (proto.Message, error) {
	vars := mux.Vars(r)
	req := &iotexapi.GetAccountRequest{
		Address: vars["addr"],
	}
	return cli.GetAccount(context.Background(), req)
}
