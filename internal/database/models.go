package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
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

const (
	KindHTTP Kind = "http"
	KindSQL  Kind = "sql"
	KindGRPC Kind = "grpc"
)

type RequestData interface {
	isRequestData() Kind
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

var decoderRequestSQL = json2.Map3(
	func(dsn string, database Database, query string) SQLRequest {
		return SQLRequest{dsn, database, query}
	},
	json2.Optional("dsn", json2.String, ""),
	json2.Map(
		json2.Optional("database", json2.String, ""),
		func(s string) Database {
			return Database(s)
		}),
	json2.Field("query", json2.String),
)

var decoderRequestGRPC = json2.Map4(
	func(target, method, payload string, metadata []KV) GRPCRequest {
		return GRPCRequest{target, method, payload, metadata}
	},
	json2.Optional("target", json2.String, ""),
	json2.Optional("method", json2.String, ""),
	json2.Optional("payload", json2.String, "{}"),
	json2.Optional("metadata", decoderKVs, nil),
)

var decoderKind = json2.Map(
	json2.Field("kind", json2.String),
	func(kind string) Kind {
		return Kind(kind)
	})

var decoderRequestData = json2.AndThen(
	decoderKind,
	func(kind Kind) json2.Decoder[RequestData] {
		switch kind {
		case KindHTTP:
			return json2.Map(decoderRequestHTTP, func(dest HTTPRequest) RequestData { return dest })
		case KindSQL:
			return json2.Map(decoderRequestSQL, func(dest SQLRequest) RequestData { return dest })
		default:
			return json2.Fail[RequestData](fmt.Sprintf("unknown request kind %q", kind))
		}
	})

type ResponseData interface {
	isResponseData() Kind
}

func ResponseDataWithKind(resp ResponseData) (map[string]any, error) {
	b, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	m["kind"] = resp.isResponseData()
	return m, nil
}

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

type Database string

const (
	Postgres   Database = "postgres"
	MySQL      Database = "mysql"
	SQLite     Database = "sqlite"
	Clickhouse Database = "clickhouse"
)

type SQLRequest struct {
	DSN      string   `json:"dsn"`
	Database Database `json:"database"`
	Query    string   `json:"query"`
}

func (SQLRequest) isRequestData() Kind { return KindSQL }

type ColumnType string

const (
	ColumnTypeString  ColumnType = "string"
	ColumnTypeNumber  ColumnType = "number"
	ColumnTypeTime    ColumnType = "time"
	ColumnTypeBoolean ColumnType = "boolean"
)

type SQLResponse struct { // TODO: last inserted id on insert
	Columns []string     `json:"columns"`
	Types   []ColumnType `json:"types"`
	Rows    [][]any      `json:"rows"`
}

func (SQLResponse) isResponseData() Kind { return KindSQL }

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

type HistoryEntry[I RequestData, O ResponseData] struct {
	SentAt     time.Time `json:"sent_at"`
	ReceivedAt time.Time `json:"received_at"`
	Request    I         `json:"request"`
	Response   O         `json:"response"`
}

type RequestID string

type Request struct {
	ID      RequestID
	Data    RequestData
	History any // TODO: []HistoryEntry[HTTPRequest, HTTPResponse] | []HistoryEntry[SQLRequest, SQLResponse] aligned w/ Data field
}

func (e *Request) UnmarshalJSON(b []byte) error {
	var decoderHistoryEntry = json2.AndThen(
		decoderKind,
		func(kind Kind) json2.Decoder[Request] {
			var decoderRequest json2.Decoder[RequestData]
			var history any
			switch kind {
			case KindHTTP:
				decoderRequest = json2.Map(decoderRequestHTTP, func(dest HTTPRequest) RequestData { return dest })
				history = []HistoryEntry[HTTPRequest, HTTPResponse]{}
				// decoderResponse = json2.Map(json2.Map3(
				// 	func(code int, body string, headers []KV) HTTPResponse {
				// 		return HTTPResponse{code, body, headers}
				// 	},
				// 	json2.Field("code", json2.Int),
				// 	json2.Optional("body", json2.String, ""),
				// 	json2.Optional("headers", decoderKVs, nil),
				// ), func(dest HTTPResponse) ResponseData { return dest })
			case KindSQL:
				decoderRequest = json2.Map(decoderRequestSQL, func(dest SQLRequest) RequestData { return dest })
				history = []HistoryEntry[SQLRequest, SQLResponse]{}
				// decoderResponse = json2.Map(json2.Map3(
				// 	func(columns []string, types []ColumnType, rows [][]any) SQLResponse {
				// 		return SQLResponse{columns, types, rows}
				// 	},
				// 	json2.Field("columns", json2.List(json2.String)),
				// 	json2.Field("types", json2.List(json2.Map(json2.String, func(s string) ColumnType {
				// 		return ColumnType(s)
				// 	}))),
				// 	json2.Field("rows", json2.List(json2.List(func(v any, t *any) error {
				// 		*t = v
				// 		return nil
				// 	}))),
				// ), func(dest SQLResponse) ResponseData { return dest })
			case KindGRPC:
				decoderRequest = json2.Map(decoderRequestGRPC, func(dest GRPCRequest) RequestData { return dest })
				history = []HistoryEntry[GRPCRequest, GRPCResponse]{}
			default:
				return json2.Fail[Request](fmt.Sprintf("unknown kind %q", kind))
			}

			return json2.Map(
				decoderRequest,
				func(req RequestData) Request {
					return Request{"", req, history}
				},
			)
		})

	var err error
	*e, err = decoderHistoryEntry.ParseBytes(b)
	return err
}

func (e Request) MarshalJSON() ([]byte, error) {
	switch req := e.Data.(type) {
	case HTTPRequest:
		return json.Marshal(map[string]any{
			"kind":    "http",
			"id":      e.ID,
			"url":     req.URL,
			"method":  req.Method,
			"body":    req.Body,
			"headers": req.Headers,
		})
	case SQLRequest:
		return json.Marshal(map[string]any{
			"kind":     "sql",
			"id":       e.ID,
			"dsn":      req.DSN,
			"database": req.Database,
			"query":    req.Query,
		})
	case GRPCRequest:
		return json.Marshal(map[string]any{
			"kind":     "grpc",
			"id":       e.ID,
			"target":   req.Target,
			"method":   req.Method,
			"payload":  req.Payload,
			"metadata": req.Metadata,
		})
	default:
		return nil, errors.Errorf("unsupported request type %T", req)
	}
}

func (e Request) MarshalJSON2() ([]byte, error) {
	switch req := e.Data.(type) {
	case HTTPRequest:
		return json.Marshal(map[string]any{
			"kind":    "http",
			"url":     req.URL,
			"method":  req.Method,
			"body":    req.Body,
			"headers": req.Headers,
		})
	case SQLRequest:
		return json.Marshal(map[string]any{
			"kind":     "sql",
			"dsn":      req.DSN,
			"database": req.Database,
			"query":    req.Query,
		})
	case GRPCRequest:
		return json.Marshal(map[string]any{
			"kind":     "grpc",
			"target":   req.Target,
			"method":   req.Method,
			"payload":  req.Payload,
			"metadata": req.Metadata,
		})
	default:
		return nil, errors.Errorf("unsupported request type %T", req)
	}
}
