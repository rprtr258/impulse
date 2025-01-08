package database

import (
	"encoding/json"
	"fmt"
	"time"

	json2 "github.com/rprtr258/fun/exp/json"
)

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var decoderKVs = json2.Map(json2.Nullable(json2.List(json2.Map2(
	func(key string, value string) KV {
		return KV{key, value}
	},
	json2.Field("key", json2.String),
	json2.Field("value", json2.String),
))), func(m json2.Maybe[[]KV]) []KV {
	return m.Value
})

type Kind string

type RequestData interface {
	isRequestData() Kind
}

var decoderKind = json2.Map(
	json2.Field("kind", json2.String),
	func(kind string) Kind {
		return Kind(kind)
	})

func decoderRequestMap[T RequestData](dest T) RequestData    { return dest }
func decoderResponseMap[T ResponseData](dest T) ResponseData { return dest }

var decodersRequest = map[Kind]json2.Decoder[RequestData]{}
var decodersResponse = map[Kind]json2.Decoder[ResponseData]{}
var decoderRequestData = json2.AndThen(
	decoderKind,
	func(kind Kind) json2.Decoder[RequestData] {
		decoder, ok := decodersRequest[kind]
		if !ok {
			return json2.Fail[RequestData](fmt.Sprintf("unknown request kind %q", kind))
		}
		return decoder
	})

type ResponseData interface {
	isResponseData() Kind
}

type HistoryEntry struct {
	SentAt     time.Time    `json:"sent_at"`
	ReceivedAt time.Time    `json:"received_at"`
	Request    RequestData  `json:"request"`
	Response   ResponseData `json:"response"`
}

type RequestID string

type Request struct {
	ID      RequestID
	Data    RequestData
	History []HistoryEntry // TODO: []HistoryEntry[HTTPRequest, HTTPResponse] | []HistoryEntry[SQLRequest, SQLResponse] aligned w/ Data field
}

var decoderRequest = json2.AndThen(
	decoderKind,
	func(kind Kind) json2.Decoder[Request] {
		decoderRequest, ok := decodersRequest[kind]
		if !ok {
			return json2.Fail[Request](fmt.Sprintf("unknown kind %q", kind))
		}

		return json2.Map(
			decoderRequest,
			func(req RequestData) Request {
				return Request{"", req, nil}
			},
		)
	})

func (e *Request) UnmarshalJSON(b []byte) error {
	var err error
	*e, err = decoderRequest.ParseBytes(b)
	return err
}

func gavnischtsche(x any) (map[string]any, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func (e Request) MarshalJSON() ([]byte, error) {
	m, err := gavnischtsche(e.Data)
	if err != nil {
		return nil, err
	}

	m["id"] = e.ID
	m["kind"] = e.Data.isRequestData()

	return json.Marshal(m)
}

func (e Request) MarshalJSON2() ([]byte, error) {
	m, err := gavnischtsche(e.Data)
	if err != nil {
		return nil, err
	}

	m["kind"] = e.Data.isRequestData()

	return json.Marshal(m)
}

func ResponseDataWithKind(resp ResponseData) (map[string]any, error) {
	m, err := gavnischtsche(resp)
	if err != nil {
		return nil, err
	}

	m["kind"] = resp.isResponseData()

	return m, nil
}

func DecodeHistory(req RequestData, b []byte) (map[string]any, error) {
	kind := req.isRequestData()
	decoderHistory := json2.Map4(
		func(sentAt, receivedAt time.Time, request RequestData, response ResponseData) map[string]any {
			res, _ := gavnischtsche(HistoryEntry{sentAt, receivedAt, request, response})
			return res
		},
		json2.Required("sent_at", json2.Time),
		json2.Required("received_at", json2.Time),
		json2.Required("request", decodersRequest[kind]),
		json2.Required("response", decodersResponse[kind]),
	)
	return decoderHistory.ParseBytes(b)
}
