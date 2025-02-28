package app

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rprtr258/fun"

	"github.com/rprtr258/impulse/internal/database"
)

func (a *App) list(
	node database.Tree,
	requests map[string]database.Request,
) error {
	for _, requestID := range node.RequestIDs {
		request, err := database.Get(a.ctx, a.DB, requestID)
		if err != nil {
			return errors.Wrapf(err, "get request id=%q", requestID)
		}
		requests[string(requestID)] = request
	}
	for _, subtree := range node.Dirs {
		if err := a.list(subtree, requests); err != nil {
			return err
		}
	}
	return nil
}

type Tree struct {
	IDs  []string
	Dirs map[string]Tree
}

type ListResponse struct {
	Tree     Tree
	Requests map[string]database.Request
	History  []map[string]any
}

func (a *App) List() (ListResponse, error) {
	tree, err := database.List(a.ctx, a.DB)
	if err != nil {
		return ListResponse{}, errors.Wrap(err, "list requests")
	}

	requests := make(map[string]database.Request)
	if err := a.list(tree, requests); err != nil { // TODO: batch
		return ListResponse{}, errors.Wrap(err, "get requests info")
	}

	history := []map[string]any{}
	for _, req := range requests {
		history = append(history, fun.Map[map[string]any](func(h database.HistoryEntry) map[string]any {
			b, _ := json.Marshal(h)

			m, _ := database.DecodeHistory(req.Data, b)
			m["RequestId"] = req.ID
			return m
		}, req.History...)...)
	}
	slices.SortFunc(history, func(i, j map[string]any) int {
		return strings.Compare(j["sent_at"].(string), i["sent_at"].(string))
	})

	var mapper func(database.Tree) Tree
	mapper = func(tree database.Tree) Tree {
		result := make(map[string]Tree, len(tree.Dirs))
		for k, dir := range tree.Dirs {
			result[k] = mapper(dir)
		}

		return Tree{
			IDs:  fun.Map[string](func(id database.RequestID) string { return string(id) }, tree.RequestIDs...),
			Dirs: result,
		}
	}
	return ListResponse{
		Tree:     mapper(tree),
		Requests: requests,
		History:  history,
	}, nil
}

type ResponseNewRequest struct {
	ID database.RequestID `json:"id"`
}

func (a *App) Create(
	id string,
	kind database.Kind,
) (ResponseNewRequest, error) {
	var req database.RequestData
	switch kind {
	case database.KindHTTP:
		req = database.HTTPRequest{
			"",             // URL // TODO: insert last url used
			http.MethodGet, // Method
			"",             // Body
			nil,            // Headers
		}
	case database.KindSQL:
		req = database.SQLRequest{
			"",                // DSN // TODO: insert last dsn used
			database.Postgres, // Database
			"",                // Query
		}
	case database.KindGRPC:
		req = database.GRPCRequest{
			"",  // Target
			"",  // Method
			"",  // Payload
			nil, // Metadata
		}
	case database.KindJQ:
		req = database.JQRequest{
			".", // Query
			`{
	"string": "string",
	"number": 42,
	"bool": true,
	"list": [1, 2, 3],
	"null": null
}`, // JSON
		}
	case database.KindRedis:
		req = database.RedisRequest{
			"localhost:6379",
			`KEYS`,
		}
	default:
		return ResponseNewRequest{}, errors.Errorf("unknown request kind %q", kind)
	}

	requestID, err := database.Create(a.ctx, a.DB, database.PayloadRequestCreate{database.RequestID(id), req})
	if err != nil {
		return ResponseNewRequest{}, errors.Wrap(err, "error while creating request")
	}

	return ResponseNewRequest{
		ID: requestID,
	}, nil
}

func (a *App) Duplicate(id string) error {
	if err := database.Duplicate(a.ctx, a.DB, id); err != nil {
		return errors.Wrap(err, "error while duplicating")
	}

	return nil
}

func (a *App) Read(requestID string) (database.Request, error) {
	request, err := database.Get(
		a.ctx, a.DB,
		database.RequestID(requestID),
	)
	if err != nil {
		return database.Request{}, errors.Wrapf(err, "get request id=%q", requestID)
	}

	return request, nil
}

func (a *App) Update(
	requestID string,
	kind database.Kind,
	newRequestID string, // TODO: rename field
	request map[string]any,
) error {
	b, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "huita request")
	}

	var requestt database.RequestData
	switch kind {
	case database.KindHTTP:
		var req database.HTTPRequest
		if err := json.Unmarshal(b, &req); err != nil {
			return errors.Wrap(err, "huita 2 request")
		}
		requestt = req
	case database.KindSQL:
		var req database.SQLRequest
		if err := json.Unmarshal(b, &req); err != nil {
			return errors.Wrap(err, "huita 3 request")
		}
		requestt = req
	case database.KindGRPC:
		var req database.GRPCRequest
		if err := json.Unmarshal(b, &req); err != nil {
			return errors.Wrap(err, "huita 4 request")
		}
		requestt = req
	case database.KindJQ:
		var req database.JQRequest
		if err := json.Unmarshal(b, &req); err != nil {
			return errors.Wrap(err, "huita 5 request")
		}
		requestt = req
	case database.KindRedis:
		var req database.RedisRequest
		if err := json.Unmarshal(b, &req); err != nil {
			return errors.Wrap(err, "huita 6 request")
		}
		requestt = req
	default:
		return errors.Errorf("unknown request kind %q", kind)
	}

	if err := database.Update(
		a.ctx, a.DB,
		database.RequestID(requestID),
		database.Kind(kind),
		database.RequestID(newRequestID),
		requestt,
	); err != nil {
		return errors.Wrap(err, "update request")
	}

	return nil
}

func (a *App) Delete(requestID string) error {
	if err := database.Delete(
		a.ctx, a.DB,
		database.RequestID(requestID),
	); err != nil {
		return errors.Wrap(err, "delete request")
	}

	return nil
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

// Perform create a handler that performs call and save result to history
func (a *App) Perform(requestID string) (map[string]any, error) {
	request, err := database.Get(
		a.ctx, a.DB,
		database.RequestID(requestID),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "get request id=%q", requestID)
	}

	sentAt := time.Now()

	var response database.ResponseData
	switch request := request.Data.(type) {
	case database.HTTPRequest:
		response, err = a.sendHTTP(request)
		if err != nil {
			return nil, errors.Wrapf(err, "send http request id=%q", requestID)
		}
	case database.SQLRequest:
		response, err = a.sendSQL(request)
		if err != nil {
			return nil, errors.Wrapf(err, "send sql request id=%q", requestID)
		}
	case database.GRPCRequest:
		response, err = a.sendGRPC(request)
		if err != nil {
			return nil, errors.Wrapf(err, "send grpc request id=%q", requestID)
		}
	case database.JQRequest:
		response, err = sendJQ(a.ctx, request)
		if err != nil {
			return nil, errors.Wrapf(err, "send jq request id=%q", requestID)
		}
	case database.RedisRequest:
		response, err = sendRedis(a.ctx, request)
		if err != nil {
			return nil, errors.Wrapf(err, "send redis request id=%q", requestID)
		}
	default:
		return nil, errors.Errorf("unsupported request type %T", request)
	}

	receivedAt := time.Now()
	if err := database.CreateHistoryEntry(
		a.ctx, a.DB, database.RequestID(requestID),
		sentAt, receivedAt,
		request.Data, response,
	); err != nil {
		return nil, errors.Wrap(err, "insert into database")
	}

	return database.ResponseDataWithKind(response)
}
