package staking_v1

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"

	"github.com/iotexproject/pharos/handler"
)

const (
	StakingContract          = "io1xpq62aw85uqzrccg9y5hnryv8ld2nkpycc3gza"
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
	req, err := http.NewRequest("POST", MemberGateway, strings.NewReader(MemberAllCandidatesQuery))
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-iotex-client-id", "web-iopay-home")

	resp, err := http.DefaultClient.Do(req)
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
		item := Validator{
			ID:     validator.RegisteredName,
			Status: true,
			Details: StakingDetails{
				// TODO how to fetch data
				Reward: StakingReward{
					Annual: 0.0,
				},
				LockTime:      259200,
				MinimumAmount: "100000000000000000000",
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
	conn, err := handler.GrpcConnection()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer conn.Close()

	cli := iotexapi.NewAPIServiceClient(conn)
	res, err := cli.ReadContract(context.Background(), &iotexapi.ReadContractRequest{
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
		res, err = cli.ReadContract(context.Background(), &iotexapi.ReadContractRequest{
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
				dalegation := Delegation{
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
							MinimumAmount: "100000000000000000000",
						},
					},
					Value:  pygg.StakedAmount.String(),
					Status: DelegationStatusActive,
				}
				delegationsPage = append(delegationsPage, dalegation)
				break
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
