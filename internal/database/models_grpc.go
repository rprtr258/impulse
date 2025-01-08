package database

import json2 "github.com/rprtr258/fun/exp/json"

const KindGRPC Kind = "grpc"

func init() {
	decoders[KindGRPC] = json2.Map(decoderRequestGRPC, decoderRequestMap)
	histories[KindGRPC] = []HistoryEntry[GRPCRequest, GRPCResponse]{}
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
