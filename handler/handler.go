package handler

import (
	"encoding/hex"
	"net/http"
	"os"
	"strconv"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/iotexproject/iotex-antenna-go/rpcmethod"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/pkg/errors"
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

// GetAccount extracts address from http request, make gRPC call GetAccount()
func GetAccount(r *http.Request, rpc *rpcmethod.RPCMethod) (proto.Message, error) {
	vars := mux.Vars(r)
	req := &iotexapi.GetAccountRequest{
		Address: vars["addr"],
	}
	return rpc.GetAccount(req)
}

// GetActionByHash extracts hash from http request, make gRPC call GetActions()
func GetActionByHash(r *http.Request, rpc *rpcmethod.RPCMethod) (proto.Message, error) {
	vars := mux.Vars(r)
	req := &iotexapi.GetActionsRequest{
		Lookup: &iotexapi.GetActionsRequest_ByHash{
			ByHash: &iotexapi.GetActionByHashRequest{
				ActionHash: vars["hash"],
			},
		},
	}
	return rpc.GetActions(req)
}

// GetActionByHash extracts address from http request, make gRPC call GetActions()
func GetActionByAddr(r *http.Request, rpc *rpcmethod.RPCMethod) (proto.Message, error) {
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
	return rpc.GetActions(req)
}

// SendSignedActionBytes extracts signed transaction bytes from http request, make gRPC call SendAction()
func SendSignedActionBytes(r *http.Request, rpc *rpcmethod.RPCMethod) (proto.Message, error) {
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
	return rpc.SendAction(req)
}

// GetTsfInBlock make gRPC call GetRawBlocks(), and returns all transfers in the block
func GetTsfInBlock(r *http.Request, rpc *rpcmethod.RPCMethod) (proto.Message, error) {
	vars := mux.Vars(r)
	blk, err := strconv.ParseInt(vars["block"], 10, 64)
	if err != nil {
		return nil, err
	}

	req := &iotexapi.GetRawBlocksRequest{
		StartHeight:  uint64(blk),
		Count:        1,
		WithReceipts: false,
	}
	resp, err := rpc.GetRawBlocks(req)
	if err != nil {
		return nil, err
	}

	var tsf []*iotextypes.Action
	for i, e := range resp.Blocks[0].Block.Body.Actions {
		if e.Core.GetTransfer() != nil {
			tsf = append(tsf, resp.Blocks[0].Block.Body.Actions[i])
		}
	}
	resp.Blocks[0].Block.Body.Actions = nil
	resp.Blocks[0].Block.Body.Actions = tsf
	return resp.Blocks[0].Block.Body, nil
}

// GetChainMeta make gRPC call GetChainMeta(), and returns the chain meta
func GetChainMeta(r *http.Request, rpc *rpcmethod.RPCMethod) (proto.Message, error) {
	resp, err := rpc.GetChainMeta(&iotexapi.GetChainMetaRequest{})
	if err != nil {
		return nil, err
	}
	return resp.ChainMeta, nil
}

// GrpcToHttpHandler turns gRPC handler into http handler
func GrpcToHttpHandler(fn func(*http.Request, *rpcmethod.RPCMethod) (proto.Message, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// connect to gRPC endpoint
		svr, err := rpcmethod.NewRPCMethod(endpoint, enableTLS)
		if err != nil {
			http.Error(w, errors.Wrap(err, ErrGrpcConnFailed).Error(), http.StatusServiceUnavailable)
			return
		}
		defer svr.Close()

		// execute gRPC call
		res, err := fn(r, svr)
		if err != nil {
			http.Error(w, errors.Wrap(err, ErrGrpcCallFailed).Error(), http.StatusInternalServerError)
			return
		}

		// marshal to JSON and write back to HTTP client
		w.Write([]byte(convertToJSON(res)))
	}
}

func convertToJSON(pb proto.Message) string {
	marshal := &jsonpb.Marshaler{
		false,
		false,
		"",
		true,
		nil,
	}
	str, err := marshal.MarshalToString(pb)
	if err != nil {
		return err.Error()
	}
	return str
}
