package database

import (
	"encoding/base64"

	json2 "github.com/rprtr258/fun/exp/json"
)

const KindSQL Kind = "sql"

func init() {
	AllKinds = append(AllKinds, enumElem[Kind]{KindSQL, "SQL"})
	decodersRequest[KindSQL] = json2.Map(decoderRequestMap, decoderRequestSQL)
	decodersResponse[KindSQL] = json2.Map(decoderResponseMap, decoderResponseSQL)
}

var decoderRequestSQL = json2.Map3(
	func(dsn string, database Database, query string) SQLRequest {
		return SQLRequest{dsn, database, query}
	},
	json2.Optional("dsn", json2.String, ""),
	json2.Map(func(s string) Database {
		return Database(s)
	}, json2.Optional("database", json2.String, "")),
	json2.Required("query", json2.String),
)

func decoderAny(v any, dest *any) error {
	*dest = v
	return nil
}

var decoderResponseSQL = json2.Map3(
	func(columns []string, types []ColumnType, rows [][]any) SQLResponse {
		for i, columnType := range types {
			if columnType != "[]uint8" {
				continue
			}

			for j, row := range rows {
				value, ok := row[i].(string)
				if !ok {
					continue
				}

				// TODO: parse jsonb
				decoded, err := base64.StdEncoding.DecodeString(value)
				if err == nil {
					rows[j][i] = string(decoded)
				}
			}
		}
		return SQLResponse{columns, types, rows}
	},
	json2.Required("columns", json2.List(json2.String)),
	json2.Required("types", json2.List(json2.Map(func(col string) ColumnType {
		return ColumnType(col)
	}, json2.String))),
	json2.Optional("rows", json2.List(json2.List(decoderAny)), [][]any{}),
)

type Database string

const (
	Postgres   Database = "postgres"
	MySQL      Database = "mysql"
	SQLite     Database = "sqlite"
	Clickhouse Database = "clickhouse"
)

var AllDatabases = []enumElem[Database]{
	{Postgres, "POSTGRES"},
	{MySQL, "MYSQL"},
	{SQLite, "SQLITE"},
	{Clickhouse, "CLICKHOUSE"},
}

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
