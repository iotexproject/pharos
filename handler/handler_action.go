package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
)

// GetActionByHash extracts hash from http request, make gRPC call GetActions()
func GetActionByHash(r *http.Request, cli iotexapi.APIServiceClient) (proto.Message, error) {
	vars := mux.Vars(r)
	req := &iotexapi.GetActionsRequest{
		Lookup: &iotexapi.GetActionsRequest_ByHash{
			ByHash: &iotexapi.GetActionByHashRequest{
				ActionHash: vars["hash"],
			},
		},
	}
	return cli.GetActions(context.Background(), req)
}

// GetActionByHash extracts address from http request, make gRPC call GetActions()
func GetActionByAddr(r *http.Request, cli iotexapi.APIServiceClient) (proto.Message, error) {
	vars := mux.Vars(r)
	start, err := strconv.ParseInt(vars["start"], 10, 64)
	if err != nil {
		return nil, err
	}
	count, err := strconv.ParseInt(vars["count"], 10, 64)
	if err != nil {
		return nil, err
	}

	req := &iotexapi.GetActionsRequest{
		Lookup: &iotexapi.GetActionsRequest_ByAddr{
			ByAddr: &iotexapi.GetActionsByAddressRequest{
				Address: vars["addr"],
				Start:   uint64(start),
				Count:   uint64(count),
			},
		},
	}
	return cli.GetActions(context.Background(), req)
}
