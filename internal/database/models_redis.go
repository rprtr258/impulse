package database

import json2 "github.com/rprtr258/fun/exp/json"

const KindRedis Kind = "redis"

var pluginRedis = plugin[RedisRequest, RedisResponse]{
	enumElem[Kind]{KindRedis, "REDIS"},
	decoderRequestRedis,
	decoderResponseRedis,
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

func (RedisRequest) Kind() Kind { return KindRedis }

type RedisResponse struct {
	Response string `json:"response"`
}

func (RedisResponse) isResponseData() Kind { return KindRedis }
