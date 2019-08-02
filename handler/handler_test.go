package handler

import (
	"net/http"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/stretchr/testify/require"
)

func Test_Pharos(t *testing.T) {
	require := require.New(t)

	baseURL := "https://pharos.iotex.io/v1"
	addr := "io1e2nqsyt7fkpzs5x7zf2uk0jj72teu5n6aku3tr"
	resp, err := http.DefaultClient.Get(baseURL + "/accounts/" + addr)
	require.NoError(err)
	require.NotNil(resp.Body)
	defer resp.Body.Close()
	act := &iotexapi.GetAccountResponse{}
	require.NoError(jsonpb.Unmarshal(resp.Body, act))
	require.Equal(addr, act.AccountMeta.Address)
	require.True(act.AccountMeta.Nonce >= 17)
	require.True(act.AccountMeta.NumActions >= 34)

	hash := "53e729d28b0c69fc66c4317fdc6ee7af292980ce781b56b502e2ee2e0b9ca48a"
	resp, err = http.DefaultClient.Get(baseURL + "/actions/hash/" + hash)
	require.NoError(err)
	require.NotNil(resp.Body)
	defer resp.Body.Close()
	acts := &iotexapi.GetActionsResponse{}
	require.NoError(jsonpb.Unmarshal(resp.Body, acts))
	require.EqualValues(1, acts.Total)
	actInfo := acts.ActionInfo[0]
	require.Equal(hash, actInfo.ActHash)
	require.Equal("33e1d2858cec24059f22348b862a2f415a21bb14b7d96733249a12e96c542969", actInfo.BlkHash)
	require.EqualValues(222656, actInfo.BlkHeight)
	require.Equal("io1e2nqsyt7fkpzs5x7zf2uk0jj72teu5n6aku3tr", actInfo.Sender)
	require.Equal("10000000000000000", actInfo.GasFee)
	require.EqualValues(1558135580, actInfo.Timestamp.Seconds)

	resp, err = http.DefaultClient.Get(baseURL + "/actions/addr/io1e2nqsyt7fkpzs5x7zf2uk0jj72teu5n6aku3tr?count=2&start=0")
	require.NoError(err)
	require.NotNil(resp.Body)
	defer resp.Body.Close()
	acts = &iotexapi.GetActionsResponse{}
	require.NoError(jsonpb.Unmarshal(resp.Body, acts))
	require.True(acts.Total >= 47)
	actInfo = acts.ActionInfo[1]
	require.Equal("0f4e20bdc0e91e65242eb08c5475292962bf92d3d624b2bc5ae61cd6e73e8161", actInfo.ActHash)
	require.Equal("a43825aa49a4a688f136f77bcdfcdb101d41a7c9886badff57ca5c0d605f3042", actInfo.BlkHash)
	require.Equal("io17ch0jth3dxqa7w9vu05yu86mqh0n6502d92lmp", actInfo.Sender)
	require.Equal("20000000000000000", actInfo.GasFee)
	require.EqualValues(1558077250, actInfo.Timestamp.Seconds)

	resp, err = http.DefaultClient.Get(baseURL + "/transfers/block/222669")
	require.NoError(err)
	require.NotNil(resp.Body)
	defer resp.Body.Close()
	blks := &iotextypes.BlockBody{}
	require.NoError(jsonpb.Unmarshal(resp.Body, blks))
	require.Equal(1, len(blks.Actions))
	require.Equal("io1e2nqsyt7fkpzs5x7zf2uk0jj72teu5n6aku3tr", blks.Actions[0].Core.GetTransfer().Recipient)

	resp, err = http.DefaultClient.Get(baseURL + "/chainmeta")
	require.NoError(err)
	require.NotNil(resp.Body)
	defer resp.Body.Close()
	meta := &iotextypes.ChainMeta{}
	require.NoError(jsonpb.Unmarshal(resp.Body, meta))
	require.True(meta.Epoch.Num > 2433)
	require.True(meta.Epoch.GravityChainStartHeight >= 8269100)
}
