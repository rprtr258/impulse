package app

import (
	"database/sql"
	"reflect"
	"time"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"

	"github.com/rprtr258/fun"
	"github.com/rprtr258/impulse/internal/database"
)

func convertTypes(columns int, rows [][]any) []database.ColumnType {
	types := make([]database.ColumnType, columns) // TODO: fix get types
	if len(rows) > 0 {
		for i := range columns {
			for _, row := range rows {
				if row[i] == nil {
					continue
				}

				types[i] = func() database.ColumnType {
					switch row[i].(type) {
					case string:
						return database.ColumnTypeString
					case uint8, uint16, uint32, uint64, int8, int16, int32, int64:
						return database.ColumnTypeNumber
					case time.Time:
						return database.ColumnTypeTime
					case bool:
						return database.ColumnTypeBoolean
					default:
						return database.ColumnType(reflect.TypeOf(row[i]).String())
					}
				}()
				break
			}
		}
	}
	return types
}

func (a *App) sendSQLClickhouse(req database.SQLRequest) (database.SQLResponse, error) {
	opts, err := clickhouse.ParseDSN(req.DSN)
	if err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "parse DSN")
	}

	db := clickhouse.OpenDB(opts)
	defer db.Close()
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.PingContext(a.ctx); err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "ping database")
	}

	// TODO: add limit
	rows, err := db.Query(req.Query)
	if err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "query")
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "get columns")
	}

	var rowsData [][]any
	for rows.Next() {
		rowDest := make([]any, len(columns))

		dest := fun.Map[any](func(_ any, i int) any {
			return &rowDest[i]
		}, rowDest...)
		if err := rows.Scan(dest...); err != nil {
			return database.SQLResponse{}, errors.Wrap(err, "scan row")
		}

		rowsData = append(rowsData, rowDest)
	}

	return database.SQLResponse{
		columns,
		convertTypes(len(columns), rowsData),
		rowsData,
	}, nil
}

func (a *App) sendSQLSqlite(req database.SQLRequest) (database.SQLResponse, error) {
	db, err := sql.Open("sqlite", req.DSN)
	if err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "connect to database")
	}
	defer db.Close()

	if err := db.PingContext(a.ctx); err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "ping database")
	}

	// TODO: add limit
	rows, err := db.Query(req.Query)
	if err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "query")
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "get columns")
	}

	var rowsData [][]any
	for rows.Next() {
		rowDest := make([]any, len(columns))

		dest := fun.Map[any](func(_ any, i int) any {
			return &rowDest[i]
		}, rowDest...)
		if err := rows.Scan(dest...); err != nil {
			return database.SQLResponse{}, errors.Wrap(err, "scan row")
		}

		rowsData = append(rowsData, rowDest)
	}

	return database.SQLResponse{
		columns,
		convertTypes(len(columns), rowsData),
		rowsData,
	}, nil
}

func (a *App) sendSQLPostgres(req database.SQLRequest) (database.SQLResponse, error) {
	db, err := sql.Open("postgres", req.DSN)
	if err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "connect to database")
	}
	defer db.Close()

	if err := db.PingContext(a.ctx); err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "ping database")
	}

	// TODO: add limit
	rows, err := db.Query(req.Query)
	if err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "query")
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "get columns")
	}

	var rowsData [][]any
	for rows.Next() {
		rowDest := make([]any, len(columns))

		dest := fun.Map[any](func(_ any, i int) any {
			return &rowDest[i]
		}, rowDest...)
		if err := rows.Scan(dest...); err != nil {
			return database.SQLResponse{}, errors.Wrap(err, "scan row")
		}

		rowsData = append(rowsData, rowDest)
	}

	return database.SQLResponse{
		columns,
		convertTypes(len(columns), rowsData),
		rowsData,
	}, nil
}

func (a *App) sendSQL(req database.SQLRequest) (database.SQLResponse, error) {
	switch req.Database {
	case database.Postgres:
		return a.sendSQLPostgres(req)
	case database.Clickhouse:
		return a.sendSQLClickhouse(req)
	case database.SQLite:
		return a.sendSQLSqlite(req)
	default:
		return database.SQLResponse{}, errors.Errorf("unsupported database: %s", req.Database)
	}
}
