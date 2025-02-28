package database

import json2 "github.com/rprtr258/fun/exp/json"

const KindRedis Kind = "redis"

func init() {
	AllKinds = append(AllKinds, enumElem[Kind]{KindRedis, "REDIS"})
	decodersRequest[KindRedis] = json2.Map(decoderRequestMap, decoderRequestRedis)
	decodersResponse[KindRedis] = json2.Map(decoderResponseMap, decoderResponseRedis)
}

var decoderRequestRedis = json2.Map2(
	func(dsn string, query string) RedisRequest {
		return RedisRequest{dsn, query}
	},
	json2.Optional("dsn", json2.String, ""),
	json2.Required("query", json2.String),
)

var decoderResponseRedis = json2.Map(
	func(response string) RedisResponse {
		return RedisResponse{response}
	},
	json2.Required("response", json2.String),
)

type RedisRequest struct {
	DSN   string `json:"dsn"`
	Query string `json:"query"`
}

func (RedisRequest) isRequestData() Kind { return KindRedis }

type RedisResponse struct {
	Response string `json:"response"`
}

func (RedisResponse) isResponseData() Kind { return KindRedis }
