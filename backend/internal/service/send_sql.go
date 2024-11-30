package service

import (
	"context"
	"database/sql"
	"reflect"
	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/impulse-http/local-backend/internal/database"
)

func (s *Service) sendSQL(ctx context.Context, req database.SQLRequest) (database.SQLResponse, error) {
	// TODO: only req.Database="postgres" is tested
	db, err := sql.Open(string(req.Database), req.DSN)
	if err != nil {
		return database.SQLResponse{}, errors.Wrap(err, "connect to database")
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
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

		rowPtrs := make([]any, len(columns))
		for i := range rowPtrs {
			rowPtrs[i] = &rowDest[i]
		}

		if err := rows.Scan(rowPtrs...); err != nil {
			return database.SQLResponse{}, errors.Wrap(err, "scan row")
		}

		rowsData = append(rowsData, rowDest)
	}

	types := make([]database.ColumnType, len(columns)) // TODO: fix get types
	if len(rowsData) > 0 {
		n := len(columns)
		for i := range n {
			for _, row := range rowsData {
				if row[i] == nil {
					continue
				}

				switch row[i].(type) {
				case string:
					types[i] = database.ColumnTypeString
				case int64:
					types[i] = database.ColumnTypeNumber
				case time.Time:
					types[i] = database.ColumnTypeTime
				case bool:
					types[i] = database.ColumnTypeBoolean
				default:
					types[i] = database.ColumnType(reflect.TypeOf(row[i]).String())
				}
				break
			}
		}
	}

	return database.SQLResponse{
		columns,
		types,
		rowsData,
	}, nil
}
