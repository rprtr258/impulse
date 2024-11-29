package service

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"

	"github.com/impulse-http/local-backend/internal/database"
)

type ResponseNewRequest struct {
	Id      database.RequestID   `json:"id"`
	Name    string               `json:"name"`
	Request database.HTTPRequest `json:"request"`
}

func (s *Service) HandlerRequestNew(ctx context.Context, request struct {
	CollectionID string `json:"id"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
}) (ResponseNewRequest, error) {
	requestID, err := database.RequestCreate(
		ctx, s.DB,
		database.CollectionID(request.CollectionID),
		database.PayloadRequestCreate{request.Name, database.HTTPRequest{
			// request.Request.URL,
			// request.Request.Method,
			// request.Request.Body,
			// request.Request.Headers,
		}})
	if err != nil {
		return ResponseNewRequest{}, errors.Wrap(err, "error while creating request")
	}

	_ = requestID
	return ResponseNewRequest{
		// requestID,
		// request.Name,
		// request.Request,
	}, nil
}

func (s *Service) HandlerRequestGet(ctx context.Context, req struct {
	CollectionID string `json:"id"`
	RequestID    string `json:"n"`
}) (database.Request, error) {
	request, err := database.RequestGet(
		ctx, s.DB,
		database.CollectionID(req.CollectionID), database.RequestID(req.RequestID),
	)
	if err != nil {
		return database.Request{}, errors.Wrapf(err, "get request id=%q", req.RequestID)
	}

	return request, nil
}

func (s *Service) HandlerRequestUpdate(ctx context.Context, request struct {
	CollectionID string         `json:"id"`
	RequestID    string         `json:"n"`
	Kind         string         `json:"kind"`
	NewRequestID string         `json:"name"`
	Request      map[string]any `json:"request"`
}) (struct{}, error) {
	b, err := json.Marshal(request.Request)
	if err != nil {
		return struct{}{}, errors.Wrap(err, "huita request")
	}

	var requestt database.RequestData
	switch request.Kind {
	case "http":
		var req database.HTTPRequest
		if err := json.Unmarshal(b, &req); err != nil {
			return struct{}{}, errors.Wrap(err, "huita 2 request")
		}
		requestt = req
	case "sql":
		var req database.SQLRequest
		if err := json.Unmarshal(b, &req); err != nil {
			return struct{}{}, errors.Wrap(err, "huita 3 request")
		}
		requestt = req
	default:
		return struct{}{}, errors.Errorf("unknown request kind %q", request.Kind)
	}

	if err := database.RequestUpdate(
		ctx, s.DB,
		database.CollectionID(request.CollectionID),
		database.RequestID(request.RequestID),
		database.Kind(request.Kind),
		database.RequestID(request.NewRequestID),
		requestt,
	); err != nil {
		return struct{}{}, errors.Wrap(err, "update request")
	}

	return struct{}{}, nil
}

func (s *Service) HandlerRequestDelete(ctx context.Context, req struct {
	CollectionID string `json:"id"`
	RequestID    string `json:"n"`
}) (struct{}, error) {
	if err := database.RequestDelete(
		ctx, s.DB,
		database.CollectionID(req.CollectionID), database.RequestID(req.RequestID),
	); err != nil {
		return struct{}{}, errors.Wrap(err, "delete request")
	}

	return struct{}{}, nil
}

func fromKV(kvs []database.KV) http.Header {
	headers := make(http.Header, len(kvs))
	for _, kv := range kvs {
		headers.Add(kv.Key, kv.Value)
	}
	return headers
}

func toKV(headers http.Header) []database.KV {
	kvs := make([]database.KV, 0, len(headers))
	for k, vs := range headers {
		kvs = append(kvs, database.KV{
			Key:   k,
			Value: vs[0],
		})
	}
	slices.SortFunc(kvs, func(a, b database.KV) int {
		return strings.Compare(a.Key, b.Key)
	})
	return kvs
}

// HandlerRequestSend create a handler that performs call and save result to history
func (s *Service) HandlerRequestSend(ctx context.Context, req struct {
	CollectionID string `json:"id"`
	RequestID    string `json:"n"`
}) (fiber.Map, error) {
	request, err := database.RequestGet(
		ctx, s.DB,
		database.CollectionID(req.CollectionID), database.RequestID(req.RequestID),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "get request id=%q", req.RequestID)
	}

	sentAt := time.Now()

	var response database.ResponseData
	switch request := request.Data.(type) {
	case database.HTTPRequest:
		resp, err := s.sendHTTP(ctx, request)
		if err != nil {
			return nil, errors.Wrapf(err, "send request id=%q", request.URL)
		}
		response = resp

		receivedAt := time.Now()

		if err := database.HistoryEntryCreate(
			ctx, s.DB,
			database.CollectionID(req.CollectionID), database.RequestID(req.RequestID),
			database.HistoryEntry[database.HTTPRequest, database.HTTPResponse]{
				SentAt:     sentAt,
				ReceivedAt: receivedAt,
				Request:    request,
				Response:   resp,
			}); err != nil {
			return nil, errors.Wrap(err, "insert into database")
		}
	case database.SQLRequest:
		resp, err := s.sendSQL(ctx, request)
		if err != nil {
			return nil, errors.Wrapf(err, "send request id=%q", request.Query)
		}
		response = resp

		receivedAt := time.Now()

		if err := database.HistoryEntryCreate(
			ctx, s.DB,
			database.CollectionID(req.CollectionID), database.RequestID(req.RequestID),
			database.HistoryEntry[database.SQLRequest, database.SQLResponse]{
				SentAt:     sentAt,
				ReceivedAt: receivedAt,
				Request:    request,
				Response:   resp,
			}); err != nil {
			return nil, errors.Wrap(err, "insert into database")
		}
	default:
		return nil, errors.Errorf("unsupported request type %T", req)
	}

	return database.ResponseDataWithKind(response)
}
