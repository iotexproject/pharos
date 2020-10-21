package handler

import (
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
)

// GrpcToHttpHandler turns gRPC handler into http handler
func GrpcToHttpHandler(fn func(*http.Request, iotexapi.APIServiceClient) (proto.Message, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// connect to gRPC endpoint
		conn, err := GrpcConnection()
		if err != nil {
			http.Error(w, errors.Wrap(err, ErrGrpcConnFailed).Error(), http.StatusServiceUnavailable)
			return
		}
		defer conn.Close()

		// execute gRPC call
		res, err := fn(r, iotexapi.NewAPIServiceClient(conn))
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
		true,
		false,
		false,
		"",
		nil,
	}
	str, err := marshal.MarshalToString(pb)
	if err != nil {
		return err.Error()
	}
	return str
}
