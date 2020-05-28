package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
)

// GetVotesByAddr extracts address from http request, make gRPC call GetBucket()
func GetVotesByAddr(r *http.Request, cli iotexapi.APIServiceClient) (proto.Message, error) {
	method := &iotexapi.ReadStakingDataMethod{
		Method: iotexapi.ReadStakingDataMethod_BUCKETS_BY_VOTER,
	}
	methodData, err := proto.Marshal(method)
	if err != nil {
		return nil, err
	}
	vars := mux.Vars(r)
	readStakingdataRequest := &iotexapi.ReadStakingDataRequest{
		Request: &iotexapi.ReadStakingDataRequest_BucketsByVoter{
			BucketsByVoter: &iotexapi.ReadStakingDataRequest_VoteBucketsByVoter{
				VoterAddress: vars["addr"],
				Pagination: &iotexapi.PaginationParam{
					Offset: uint32(0),
					Limit:  uint32(1000),
				},
			},
		},
	}
	requestData, err := proto.Marshal(readStakingdataRequest)
	if err != nil {
		return nil, err
	}
	request := &iotexapi.ReadStateRequest{
		ProtocolID: []byte("staking"),
		MethodName: methodData,
		Arguments:  [][]byte{requestData},
	}

	response, err := cli.ReadState(context.Background(), request)
	if err != nil {
		return nil, err
	}

	bucketlist := &iotextypes.VoteBucketList{}
	if err := proto.Unmarshal(response.Data, bucketlist); err != nil {
		return nil, err
	}
	return bucketlist, nil
}

// GetVotesByIndex extracts address from http request, make gRPC call GetBucket()
func GetVotesByIndex(r *http.Request, cli iotexapi.APIServiceClient) (proto.Message, error) {
	vars := mux.Vars(r)
	index, err := strconv.ParseInt(vars["index"], 10, 64)
	if err != nil {
		return nil, err
	}
	method := &iotexapi.ReadStakingDataMethod{
		Method: iotexapi.ReadStakingDataMethod_BUCKETS_BY_INDEXES,
	}
	methodData, err := proto.Marshal(method)
	if err != nil {
		return nil, err
	}
	readStakingdataRequest := &iotexapi.ReadStakingDataRequest{
		Request: &iotexapi.ReadStakingDataRequest_BucketsByIndexes{
			BucketsByIndexes: &iotexapi.ReadStakingDataRequest_VoteBucketsByIndexes{
				Index: []uint64{uint64(index)},
			},
		},
	}
	requestData, err := proto.Marshal(readStakingdataRequest)
	if err != nil {
		return nil, err
	}
	request := &iotexapi.ReadStateRequest{
		ProtocolID: []byte("staking"),
		MethodName: methodData,
		Arguments:  [][]byte{requestData},
	}

	response, err := cli.ReadState(context.Background(), request)
	if err != nil {
		return nil, err
	}

	bucketlist := &iotextypes.VoteBucketList{}
	if err := proto.Unmarshal(response.Data, bucketlist); err != nil {
		return nil, err
	}
	return bucketlist.Buckets[0], nil
}
