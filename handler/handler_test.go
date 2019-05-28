package handler

import (
	"net/http"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
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
	require.Equal(uint64(1), acts.Total)
	actInfo := acts.ActionInfo[0]
	require.Equal(hash, actInfo.ActHash)
	require.Equal("33e1d2858cec24059f22348b862a2f415a21bb14b7d96733249a12e96c542969", actInfo.BlkHash)
	require.Equal(uint64(222656), actInfo.BlkHeight)
	require.Equal("io1e2nqsyt7fkpzs5x7zf2uk0jj72teu5n6aku3tr", actInfo.Sender)
	require.Equal("10000000000000000", actInfo.GasFee)
	require.Equal(int64(1558135580), actInfo.Timestamp.Seconds)

	resp, err = http.DefaultClient.Get(baseURL + "/actions/addr/io1e2nqsyt7fkpzs5x7zf2uk0jj72teu5n6aku3tr?count=2&start=0")
	require.NoError(err)
	require.NotNil(resp.Body)
	defer resp.Body.Close()
	acts = &iotexapi.GetActionsResponse{}
	require.NoError(jsonpb.Unmarshal(resp.Body, acts))
	require.Equal(uint64(34), acts.Total)
	actInfo = acts.ActionInfo[1]
	require.Equal("fa8faa5524e5e9c7891514fbbe3c16ffd28f42bd945858533fd0b5287083faee", actInfo.ActHash)
	require.Equal("9c41f01ce090927df0e9e4669a110555f8f918f76884e16d2939354876e2d57b", actInfo.BlkHash)
	require.Equal(uint64(222669), actInfo.BlkHeight)
	require.Equal("io1e2nqsyt7fkpzs5x7zf2uk0jj72teu5n6aku3tr", actInfo.Sender)
	require.Equal("10000000000000000", actInfo.GasFee)
	require.Equal(int64(1558135710), actInfo.Timestamp.Seconds)
}
