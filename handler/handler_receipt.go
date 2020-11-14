package handler

import (
	"context"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
)

// GetReceiptByHash extracts hash from http request, make gRPC call GetActions()
func GetReceiptByHash(r *http.Request, cli iotexapi.APIServiceClient) (proto.Message, error) {
	vars := mux.Vars(r)
	return cli.GetReceiptByAction(context.Background(), &iotexapi.GetReceiptByActionRequest{
		ActionHash: vars["hash"],
	})
}
