package database

import (
	json2 "github.com/rprtr258/fun/exp/json"
)

const KindGRPC Kind = "grpc"

func init() {
	decodersRequest[KindGRPC] = json2.Map(decoderRequestGRPC, decoderRequestMap)
	decodersResponse[KindGRPC] = json2.Map(decoderResponseGRPC, decoderResponseMap)
}

var decoderRequestGRPC = json2.Map4(
	func(target, method, payload string, metadata []KV) GRPCRequest {
		return GRPCRequest{target, method, payload, metadata}
	},
	json2.Optional("target", json2.String, ""),
	json2.Optional("method", json2.String, ""),
	json2.Optional("payload", json2.String, "{}"),
	json2.Optional("metadata", decoderKVs, nil),
)

var decoderResponseGRPC = json2.Map3(
	func(response string, code int, metadata []KV) GRPCResponse {
		return GRPCResponse{response, code, metadata}
	},
	json2.Required("response", json2.String),
	json2.Required("code", json2.Int),
	json2.Optional("metadata", decoderKVs, nil),
)

type GRPCRequest struct {
	Target   string `json:"target"`
	Method   string `json:"method"` // NOTE: fully qualified
	Payload  string `json:"payload"`
	Metadata []KV   `json:"metadata"`
}

func (GRPCRequest) isRequestData() Kind { return KindGRPC }

type GRPCResponse struct { // TODO: last inserted id on insert
	Response string `json:"response"`
	// https://grpc.io/docs/guides/status-codes/#the-full-list-of-status-codes
	Code     int  `json:"code"`
	Metadata []KV `json:"metadata"`
}

func (GRPCResponse) isResponseData() Kind { return KindGRPC }
