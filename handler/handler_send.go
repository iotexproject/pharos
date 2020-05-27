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

// SendSignedActionBytes extracts signed transaction bytes from http request, make gRPC call SendAction()
func SendSignedActionBytes(r *http.Request, cli iotexapi.APIServiceClient) (proto.Message, error) {
	vars := mux.Vars(r)
	// input is hex string of signed action bytes
	actionBytes, err := hex.DecodeString(vars["signedbytes"])
	if err != nil {
		return nil, err
	}

	action := &iotextypes.Action{}
	if err := proto.Unmarshal(actionBytes, action); err != nil {
		return nil, err
	}

	req := &iotexapi.SendActionRequest{
		Action: action,
	}
	return cli.SendAction(context.Background(), req)
}
