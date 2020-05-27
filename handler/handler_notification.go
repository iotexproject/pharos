package handler

import (
	"context"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"

	"github.com/iotexproject/go-pkgs/crypto"
	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
)

// GetTsfInBlock make gRPC call GetRawBlocks(), and returns all transfers in the block
func GetTsfInBlock(r *http.Request, cli iotexapi.APIServiceClient) (proto.Message, error) {
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
	resp, err := cli.GetRawBlocks(context.Background(), req)
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
