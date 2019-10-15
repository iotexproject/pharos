package handler

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-antenna-go/rpcmethod"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/pkg/errors"
)

const (
	// TODO change to mainnet contract address
	StakingContract          = "io1w97pslyg7qdayp8mfnffxkjkpapaf83wmmll2l"
	MemberGateway            = "https://member.iotex.io/api-gateway/"
	MemberAllCandidatesQuery = `{"operationName":"bpCandidates","variables":{},"query":"query bpCandidates {bpCandidates {id rank logo name status category serverStatus liveVotes liveVotesDelta percent registeredName socialMedia productivity productivityBase __typename}}"}`
)

type pygg struct {
	CanName          [12]byte
	StakedAmount     *big.Int
	StakeDuration    *big.Int
	StakeStartTime   *big.Int
	NonDecay         bool
	UnstakeStartTime *big.Int
	PyggOwner        common.Address
	CreateTime       *big.Int
	Prev             *big.Int
	Next             *big.Int
}

func getMemberDelegates() (*MemberDelegates, error) {
	resp, err := http.Post(MemberGateway, "application/json", strings.NewReader(MemberAllCandidatesQuery))

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
	for _, validator := range memberDelegates.Data.BPCandidates {
		var item = Validator{
			ID:     validator.RegisteredName,
			Status: true,
			Details: StakingDetails{
				// TODO how to fetch data
				Reward: StakingReward{
					Annual: 0.0,
				},
				LockTime:      259200,
				MinimumAmount: "1200000",
			},
		}
		validatePage = append(validatePage, item)
	}
	body, err := json.Marshal(validatePage)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(body)
}

func MemberDelegations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	abi, err := abi.JSON(strings.NewReader(StakingABI))
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	addr, _ := address.FromString(vars["addr"])
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	var addrArray [20]byte
	copy(addrArray[:], addr.Bytes()[:])

	data, err := abi.Pack("getPyggIndexesByAddress", addrArray)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	// connect to gRPC endpoint
	svr, err := rpcmethod.NewRPCMethod(endpoint, enableTLS)
	if err != nil {
		http.Error(w, errors.Wrap(err, ErrGrpcConnFailed).Error(), http.StatusServiceUnavailable)
		return
	}
	defer svr.Close()
	res, err := svr.ReadContract(&iotexapi.ReadContractRequest{
		Execution: &iotextypes.Execution{
			Amount:   "0",
			Contract: StakingContract,
			Data:     data,
		},
		CallerAddress: addr.String(),
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	var indexes []*big.Int
	err = abi.Unpack(&indexes, "getPyggIndexesByAddress", common.Hex2Bytes(res.Data))
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	var pyggs []pygg
	for _, index := range indexes {
		data, err = abi.Pack("pyggs", index)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		res, err = svr.ReadContract(&iotexapi.ReadContractRequest{
			Execution: &iotextypes.Execution{
				Amount:   "0",
				Contract: StakingContract,
				Data:     data,
			},
			CallerAddress: addr.String(),
		})
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		var pygg pygg
		err = abi.Unpack(&pygg, "pyggs", common.Hex2Bytes(res.Data))
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		pyggs = append(pyggs, pygg)
	}

	memberDelegates, err := getMemberDelegates()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	delegationsPage := make(DelegationsPage, 0)
	for _, pygg := range pyggs {
		name := string(pygg.CanName[:])
		for _, validator := range memberDelegates.Data.BPCandidates {
			if name == validator.RegisteredName {
				var dalegation = Delegation{
					Delegator: StakeValidator{
						ID:     validator.RegisteredName,
						Status: true,
						Info: StakeValidatorInfo{
							Name:        validator.Name,
							Description: "",
							Image:       validator.Logo,
							Website: func() string {
								if len(validator.SocialMedia) > 0 {
									return validator.SocialMedia[0]
								}
								return ""
							}(),
						},
						Details: StakingDetails{
							// TODO how to fetch data
							Reward: StakingReward{
								Annual: 0.0,
							},
							LockTime:      259200,
							MinimumAmount: "1200000",
						},
					},
					Value:  pygg.StakedAmount.String(),
					Status: DelegationStatusActive,
				}
				delegationsPage = append(delegationsPage, dalegation)
			}
		}
	}
	body, err := json.Marshal(delegationsPage)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(body)
}
