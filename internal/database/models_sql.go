package database

import json2 "github.com/rprtr258/fun/exp/json"

const KindSQL Kind = "sql"

func init() {
	decodersRequest[KindSQL] = json2.Map(decoderRequestSQL, decoderRequestMap)
	decodersResponse[KindSQL] = json2.Map(decoderResponseSQL, decoderResponseMap)
}

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

func decoderAny(v any, dest *any) error {
	*dest = v
	return nil
}

var decoderResponseSQL = json2.Map3(
	func(columns []string, types []ColumnType, rows [][]any) SQLResponse {
		return SQLResponse{columns, types, rows}
	},
	json2.Required("columns", json2.List(json2.String)),
	json2.Required("types", json2.List(json2.Map(
		json2.String,
		func(col string) ColumnType {
			return ColumnType(col)
		}))),
	json2.Required("rows", json2.List(json2.List(decoderAny))),
)

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
