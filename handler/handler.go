package handler

import (
	"encoding/hex"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/iotexproject/go-pkgs/crypto"
	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-antenna-go/rpcmethod"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	enableTLS = tls == "TRUE" || tls == "True" || endpoint == Mainnet
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
	height, err := strconv.ParseInt(vars["block"], 10, 64)
	if err != nil {
		return nil, err
	}

	req := &iotexapi.GetRawBlocksRequest{
		StartHeight:  uint64(height),
		Count:        1,
		WithReceipts: false,
	}
	resp, err := rpc.GetRawBlocks(req)
	if err != nil {
		return nil, err
	}

	blk := resp.Blocks[0].Block
	var tsf []*iotextypes.Action
	for i, e := range blk.Body.Actions {
		if e.Core.GetTransfer() != nil {
			tsf = append(tsf, blk.Body.Actions[i])
		}
	}
	// convert to ActionInfo
	var action iotexapi.GetActionsResponse
	if len(tsf) == 0 {
		return &action, nil
	}

	action.Total = uint64(len(tsf))
	action.ActionInfo = make([]*iotexapi.ActionInfo, len(tsf))
	ser, err := proto.Marshal(blk.Header)
	if err != nil {
		return nil, err
	}
	blkHash := hash.Hash256b(ser)
	for i := 0; i < len(tsf); i++ {
		ser, err := proto.Marshal(tsf[i])
		if err != nil {
			return nil, err
		}
		actHash := hash.Hash256b(ser)
		pk, err := crypto.BytesToPublicKey(tsf[i].SenderPubKey)
		if err != nil {
			return nil, err
		}
		sender, err := address.FromBytes(pk.Hash())
		if err != nil {
			return nil, err
		}
		action.ActionInfo[i] = &iotexapi.ActionInfo{
			Action:    tsf[i],
			ActHash:   hex.EncodeToString(actHash[:]),
			BlkHash:   hex.EncodeToString(blkHash[:]),
			BlkHeight: uint64(height),
			Sender:    sender.String(),
			Timestamp: blk.Header.Core.Timestamp,
		}
	}
	return &action, nil
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

func getMemberDelegates() (*MemberDelegates, error) {
	resp, err := http.Post("https://member.iotex.io/api-gateway/", "application/json",
		strings.NewReader(`{"operationName":"bpCandidates","variables":{},"query":"query bpCandidates {bpCandidates {id rank logo name status category serverStatus liveVotes liveVotesDelta percent registeredName socialMedia productivity productivityBase __typename}}"}`))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var memberDelegates MemberDelegates
	err = json.Unmarshal(body, &memberDelegates)
	return &memberDelegates, err
}

func MemberValidators(w http.ResponseWriter, r *http.Request) {
	memberDelegates, err := getMemberDelegates()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	validatePage := make(ValidatorPage, 0)
	for _, delegate := range memberDelegates.Data.BPCandidates {
		if delegate.Status == "ELECTED" {
			var validator = Validator{
				ID:     delegate.ID,
				Status: true,
				Details: StakingDetails{
					// TODO how to fetch data
					Reward: StakingReward{
						Annual: 0.0,
					},
					LockTime:      0,
					MinimumAmount: "1000000000000",
				},
			}
			validatePage = append(validatePage, validator)
		}
	}
	body, err := json.Marshal(validatePage)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(body)
}

func MemberDelegations(w http.ResponseWriter, r *http.Request) {
	memberDelegates, err := getMemberDelegates()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	delegationsPage := make(DelegationsPage, 0)
	for _, delegate := range memberDelegates.Data.BPCandidates {
		var dalegation = Delegation{
			Delegator: StakeValidator{
				ID:     delegate.ID,
				Status: true,
				Info: StakeValidatorInfo{
					Name:        delegate.Name,
					Description: "",
					Image:       delegate.Logo,
					Website: func() string {
						if len(delegate.SocialMedia) > 0 {
							return delegate.SocialMedia[0]
						}
						return ""
					}(),
				},
				Details: StakingDetails{
					// TODO how to fetch data
					Reward: StakingReward{
						Annual: 0.0,
					},
					LockTime:      0,
					MinimumAmount: "1000000000000",
				},
			},
			Value:  strconv.FormatInt(delegate.LiveVotes, 10),
			Status: DelegationStatusActive,
		}
		delegationsPage = append(delegationsPage, dalegation)
	}
	body, err := json.Marshal(delegationsPage)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(body)
}
