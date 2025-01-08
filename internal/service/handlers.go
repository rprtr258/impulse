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
	"github.com/rprtr258/fun"

	"github.com/impulse-http/local-backend/internal/database"
)

func (s *Service) list(
	ctx context.Context,
	node database.Tree,
	base string,
	requests map[string]database.Request,
) error {
	for _, requestID := range node.RequestIDs {
		request, err := database.Get(ctx, s.DB, database.RequestID(base+string(requestID)))
		if err != nil {
			return errors.Wrapf(err, "get request id=%q", requestID)
		}
		requests[base+string(requestID)] = request
	}
	for dir, subtree := range node.Dirs {
		if err := s.list(ctx, subtree, base+dir+"/", requests); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) HandlerList(ctx context.Context, _ struct{}) (fiber.Map, error) {
	tree, err := database.List(ctx, s.DB)
	if err != nil {
		return nil, errors.Wrap(err, "list requests")
	}

	requests := make(map[string]database.Request)
	if err := s.list(ctx, tree, "", requests); err != nil { // TODO: batch
		return nil, errors.Wrap(err, "get requests info")
	}

	history := []fiber.Map{}
	for _, req := range requests {
		switch req.Data.(type) {
		case database.HTTPRequest:
			history = append(history, fun.Map[fiber.Map](func(h database.HistoryEntry[database.HTTPRequest, database.HTTPResponse]) fiber.Map {
				b, _ := json.Marshal(h)

				var m fiber.Map
				_ = json.Unmarshal(b, &m)

				m["request_id"] = req.ID
				return m
			}, req.History.([]database.HistoryEntry[database.HTTPRequest, database.HTTPResponse])...)...)
		case database.SQLRequest:
			history = append(history, fun.Map[fiber.Map](func(h database.HistoryEntry[database.SQLRequest, database.SQLResponse]) fiber.Map {
				b, _ := json.Marshal(h)

				var m fiber.Map
				_ = json.Unmarshal(b, &m)

				m["request_id"] = req.ID
				return m
			}, req.History.([]database.HistoryEntry[database.SQLRequest, database.SQLResponse])...)...)
		case database.GRPCRequest:
			history = append(history, fun.Map[fiber.Map](func(h database.HistoryEntry[database.GRPCRequest, database.GRPCResponse]) fiber.Map {
				b, _ := json.Marshal(h)

				var m fiber.Map
				_ = json.Unmarshal(b, &m)

				m["request_id"] = req.ID
				return m
			}, req.History.([]database.HistoryEntry[database.GRPCRequest, database.GRPCResponse])...)...)
		case database.JQRequest:
			history = append(history, fun.Map[fiber.Map](func(h database.HistoryEntry[database.JQRequest, database.JQResponse]) fiber.Map {
				b, _ := json.Marshal(h)

				var m fiber.Map
				_ = json.Unmarshal(b, &m)

				m["request_id"] = req.ID
				return m
			}, req.History.([]database.HistoryEntry[database.JQRequest, database.JQResponse])...)...)
		default:
			return nil, errors.Errorf("unknown request type %T", req)
		}
	}
	slices.SortFunc(history, func(i, j fiber.Map) int {
		return strings.Compare(j["sent_at"].(string), i["sent_at"].(string))
	})

	var mapper func(database.Tree, string) fiber.Map
	mapper = func(tree database.Tree, base string) fiber.Map {
		result := make(map[string]fiber.Map, len(tree.Dirs))
		for k, dir := range tree.Dirs {
			result[k] = mapper(dir, base+k+"/")
		}

		return fiber.Map{
			"ids":  fun.Map[string](func(id database.RequestID) string { return base + string(id) }, tree.RequestIDs...),
			"dirs": result,
		}
	}
	return fiber.Map{
		"tree":     mapper(tree, ""),
		"requests": requests,
		"history":  history,
	}, nil
}

type ResponseNewRequest struct {
	Id      database.RequestID   `json:"id"`
	Name    string               `json:"name"`
	Request database.HTTPRequest `json:"request"`
}

func (s *Service) HandlerNew(ctx context.Context, request struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}) (ResponseNewRequest, error) {
	var req database.RequestData
	switch request.Kind {
	case "http":
		req = database.HTTPRequest{
			"",             // URL // TODO: insert last url used
			http.MethodGet, // Method
			"",             // Body
			nil,            // Headers
		}
	case "sql":
		req = database.SQLRequest{
			"",                // DSN // TODO: insert last dsn used
			database.Postgres, // Database
			"",                // Query
		}
	default:
		return ResponseNewRequest{}, errors.Errorf("unknown request kind %q", request.Kind)
	}

	requestID, err := database.Create(ctx, s.DB, database.PayloadRequestCreate{database.RequestID(request.ID), req})
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

func (s *Service) HandlerGet(ctx context.Context, req struct {
	RequestID string `json:"id"`
}) (database.Request, error) {
	request, err := database.Get(
		ctx, s.DB,
		database.RequestID(req.RequestID),
	)
	if err != nil {
		return database.Request{}, errors.Wrapf(err, "get request id=%q", req.RequestID)
	}

	return request, nil
}

func (s *Service) HandlerUpdate(ctx context.Context, request struct {
	RequestID    string         `json:"id"`
	Kind         string         `json:"kind"`
	NewRequestID string         `json:"name"` // TODO: rename field
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
	case "grpc":
		var req database.GRPCRequest
		if err := json.Unmarshal(b, &req); err != nil {
			return struct{}{}, errors.Wrap(err, "huita 4 request")
		}
		requestt = req
	case "jq":
		var req database.JQRequest
		if err := json.Unmarshal(b, &req); err != nil {
			return struct{}{}, errors.Wrap(err, "huita 5 request")
		}
		requestt = req
	default:
		return struct{}{}, errors.Errorf("unknown request kind %q", request.Kind)
	}

	if err := database.Update(
		ctx, s.DB,
		database.RequestID(request.RequestID),
		database.Kind(request.Kind),
		database.RequestID(request.NewRequestID),
		requestt,
	); err != nil {
		return struct{}{}, errors.Wrap(err, "update request")
	}

	return struct{}{}, nil
}

func (s *Service) HandlerDelete(ctx context.Context, req struct {
	RequestID string `json:"id"`
}) (struct{}, error) {
	if err := database.Delete(
		ctx, s.DB,
		database.RequestID(req.RequestID),
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

// HandlerSend create a handler that performs call and save result to history
func (s *Service) HandlerSend(ctx context.Context, req struct {
	RequestID string `json:"id"`
}) (fiber.Map, error) {
	request, err := database.Get(
		ctx, s.DB,
		database.RequestID(req.RequestID),
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
			return nil, errors.Wrapf(err, "send request id=%q", req.RequestID)
		}
		response = resp

		receivedAt := time.Now()

		if err := database.CreateHistoryEntry(
			ctx, s.DB,
			database.RequestID(req.RequestID),
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
			return nil, errors.Wrapf(err, "send request id=%q", req.RequestID)
		}
		response = resp

		receivedAt := time.Now()

		if err := database.CreateHistoryEntry(
			ctx, s.DB,
			database.RequestID(req.RequestID),
			database.HistoryEntry[database.SQLRequest, database.SQLResponse]{
				SentAt:     sentAt,
				ReceivedAt: receivedAt,
				Request:    request,
				Response:   resp,
			}); err != nil {
			return nil, errors.Wrap(err, "insert into database")
		}
	case database.GRPCRequest:
		resp, err := s.sendGRPC(ctx, request)
		if err != nil {
			return nil, errors.Wrapf(err, "send request id=%q", req.RequestID)
		}
		response = resp

		receivedAt := time.Now()

		if err := database.CreateHistoryEntry(
			ctx, s.DB,
			database.RequestID(req.RequestID),
			database.HistoryEntry[database.GRPCRequest, database.GRPCResponse]{
				SentAt:     sentAt,
				ReceivedAt: receivedAt,
				Request:    request,
				Response:   resp,
			}); err != nil {
			return nil, errors.Wrap(err, "insert into database")
		}
	case database.JQRequest:
		resps := []string{}
		for _, json := range request.JSON {
			resp, err := jq(ctx, json, request.Query)
			if err != nil {
				return nil, errors.Wrapf(err, "send request id=%q", req.RequestID)
			}
			resps = append(resps, resp...)
		}
		response = database.JQResponse{
			Response: resps,
		}

		receivedAt := time.Now()

		if err := database.CreateHistoryEntry(
			ctx, s.DB,
			database.RequestID(req.RequestID),
			database.HistoryEntry[database.JQRequest, database.JQResponse]{
				SentAt:     sentAt,
				ReceivedAt: receivedAt,
				Request:    request,
				Response:   response.(database.JQResponse),
			}); err != nil {
			return nil, errors.Wrap(err, "insert into database")
		}
	default:
		return nil, errors.Errorf("unsupported request type %T", req)
	}

	return database.ResponseDataWithKind(response)
}
