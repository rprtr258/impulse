package database

import json2 "github.com/rprtr258/fun/exp/json"

const KindJQ Kind = "jq"

func init() {
	decodersRequest[KindJQ] = json2.Map(decoderRequestJQ, decoderRequestMap)
	decodersResponse[KindJQ] = json2.Map(decoderResponseJQ, decoderResponseMap)
}

var decoderRequestJQ = json2.Map2(
	func(query string, json []string) JQRequest {
		return JQRequest{query, json}
	},
	json2.Optional("query", json2.String, "."),
	json2.Required("json", json2.List(json2.String)),
)

var decoderResponseJQ = json2.Map(
	json2.Required("response", json2.List(json2.String)),
	func(response []string) JQResponse {
		return JQResponse{response}
	},
)

type JQRequest struct {
	Query string   `json:"query"`
	JSON  []string `json:"json"`
}

func (JQRequest) isRequestData() Kind { return KindJQ }

type JQResponse struct {
	Response []string `json:"response"`
}

func (JQResponse) isResponseData() Kind { return KindJQ }
