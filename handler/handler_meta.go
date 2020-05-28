package handler

import (
	"context"
	"net/http"

	"github.com/golang/protobuf/proto"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
)

// GetChainMeta make gRPC call GetChainMeta(), and returns the chain meta
func GetChainMeta(r *http.Request, cli iotexapi.APIServiceClient) (proto.Message, error) {
	resp, err := cli.GetChainMeta(context.Background(), &iotexapi.GetChainMetaRequest{})
	if err != nil {
		return nil, err
	}
	return resp.ChainMeta, nil
}
