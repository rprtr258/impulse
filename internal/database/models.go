package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rprtr258/fun"
	json2 "github.com/rprtr258/fun/exp/json"
)

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var decoderKVs = json2.Map(func(m fun.Option[[]KV]) []KV {
	return m.Value
}, json2.Nullable(json2.List(json2.Map2(
	func(key string, value string) KV {
		return KV{key, value}
	},
	json2.Field("key", json2.String),
	json2.Field("value", json2.String),
))))

type plugin[Req RequestData, Resp ResponseData] struct {
	kind            enumElem[Kind]
	decoderRequest  json2.Decoder[Req]
	decoderResponse json2.Decoder[Resp]
}

var plugins = map[Kind]plugin[RequestData, ResponseData]{}
var AllKinds []enumElem[Kind]

func usePlugin[Req RequestData, Resp ResponseData](plug plugin[Req, Resp]) {
	plugins[plug.kind.Value] = plugin[RequestData, ResponseData]{
		kind:            plug.kind,
		decoderRequest:  json2.Map(decoderRequestMap, plug.decoderRequest),
		decoderResponse: json2.Map(decoderResponseMap, plug.decoderResponse),
	}
}

func init() {
	usePlugin(pluginRedis)
	usePlugin(pluginSql)
	usePlugin(pluginJQ)
	usePlugin(pluginGRPC)
	usePlugin(pluginHTTP)

	for _, plugin := range plugins {
		AllKinds = append(AllKinds, plugin.kind)
	}
}

type Kind string

type RequestData interface {
	Kind() Kind
}

var decoderKind = json2.Map(func(kind string) Kind {
	return Kind(kind)
}, json2.Field("kind", json2.String))

func decoderRequestMap[T RequestData](dest T) RequestData    { return dest }
func decoderResponseMap[T ResponseData](dest T) ResponseData { return dest }

type enumElem[T any] struct {
	Value  T
	TSName string
}

var decoderRequestData = json2.AndThen(decoderKind, func(kind Kind) json2.Decoder[RequestData] {
	plugin, ok := plugins[kind]
	if !ok {
		return json2.Fail[RequestData](fmt.Sprintf("unknown request kind %q", kind))
	}
	return plugin.decoderRequest
})
var decoderRequest = json2.AndThen(decoderKind, func(kind Kind) json2.Decoder[Request] {
	return json2.Map(func(req RequestData) Request {
		return Request{"", req, nil}
	}, decoderRequestData)
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
	m["kind"] = e.Data.Kind()

	return json.Marshal(m)
}

func (e Request) MarshalJSON2() ([]byte, error) {
	m, err := gavnischtsche(e.Data)
	if err != nil {
		return nil, err
	}

	m["kind"] = e.Data.Kind()

	return json.Marshal(m)
}

func DecodeHistory(req RequestData) json2.Decoder[HistoryEntry] {
	kind := req.Kind()
	plugin := plugins[kind]
	return json2.Map4(
		func(sentAt, receivedAt time.Time, request RequestData, response ResponseData) HistoryEntry {
			return HistoryEntry{sentAt, receivedAt, request, response}
		},
		json2.Required("sent_at", json2.Time),
		json2.Required("received_at", json2.Time),
		json2.Required("request", plugin.decoderRequest),
		json2.Required("response", plugin.decoderResponse),
	)
}
