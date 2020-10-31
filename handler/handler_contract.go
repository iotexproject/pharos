package handler

import (
	"context"
	"encoding/hex"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
)

// GetActionByHash extracts address from http request, make gRPC call GetActions()
func ReadContract(r *http.Request, cli iotexapi.APIServiceClient) (proto.Message, error) {
	vars := mux.Vars(r)
	method, err := hex.DecodeString(vars["method"])
	if err != nil {
		return nil, err
	}
	data, err := hex.DecodeString(vars["data"])
	if err != nil {
		return nil, err
	}

	req := &iotexapi.ReadContractRequest{
		Execution: &iotextypes.Execution{
			Amount: "0",
			Contract: vars["addr"],
			Data: append(method, data...),
		},
		CallerAddress: "io1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqd39ym7",
	}
	res, err := cli.ReadContract(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// we only care res.Data, get rid of receipt
	res.Receipt = nil
	return res, nil
}
