package database

import json2 "github.com/rprtr258/fun/exp/json"

const KindHTTP Kind = "http"

func init() {
	decoders[KindHTTP] = json2.Map(decoderRequestHTTP, decoderRequestMap)
	histories[KindHTTP] = []HistoryEntry[HTTPRequest, HTTPResponse]{}
}

var decoderRequestHTTP = json2.Map4(
	func(url string, method string, body string, headers []KV) HTTPRequest {
		return HTTPRequest{url, method, body, headers}
	},
	json2.Optional("url", json2.String, ""),
	json2.Optional("method", json2.String, "GET"),
	json2.Optional("body", json2.String, ""),
	json2.Optional("headers", decoderKVs, nil),
)

type HTTPRequest struct {
	URL     string `json:"url"`
	Method  string `json:"method"`
	Body    string `json:"body"`
	Headers []KV   `json:"headers"`
}

func (HTTPRequest) isRequestData() Kind { return KindHTTP }

type HTTPResponse struct {
	Code    int    `json:"code"`
	Body    string `json:"body"`
	Headers []KV   `json:"headers"`
}

func (HTTPResponse) isResponseData() Kind { return KindHTTP }
