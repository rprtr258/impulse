package service

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rprtr258/fun"

	"github.com/impulse-http/local-backend/internal/database"
)

type Collection struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	RequestIDs []string `json:"request_ids"`
}

func (s *Service) HandlerCollectionCreate(ctx context.Context, Request struct {
	Name string `json:"name"`
}) (fiber.Map, error) {
	col, err := database.CollectionCreate(ctx, s.DB, Request.Name)
	if err != nil {
		return nil, errors.Wrap(err, "create new collection")
	}

	return fiber.Map{
		"id": col.ID,
	}, nil
}

func (s *Service) HandlerCollectionList(ctx context.Context, request struct{}) ([]Collection, error) {
	cols, err := database.CollectionList(ctx, s.DB)
	if err != nil {
		return nil, errors.Wrap(err, "get collections list")
	}

	return fun.Map[Collection](func(col database.Collection) Collection {
		return Collection{
			ID:   string(col.ID),
			Name: col.Name,
			RequestIDs: fun.Map[string](
				func(id database.RequestID) string { return string(id) },
				col.RequestIDs...),
		}
	}, cols...), nil
}

func (s *Service) HandlerCollectionGet(ctx context.Context, request struct {
	ID string `json:"id"`
}) (fiber.Map, error) {
	collection, err := database.CollectionGet(ctx, s.DB, database.CollectionID(request.ID))
	if err != nil {
		return nil, errors.Wrap(err, "list requests")
	}

	requests := make([]database.Request, 0, len(collection.RequestIDs))
	for _, requestID := range collection.RequestIDs { // TODO: batch
		request, err := database.RequestGet(ctx, s.DB, collection.ID, requestID)
		if err != nil {
			return nil, errors.Wrapf(err, "get request id=%q", requestID)
		}
		requests = append(requests, request)
	}

	history := fun.ConcatMap(func(req database.Request) []fiber.Map {
		switch req.Data.(type) {
		case database.HTTPRequest:
			return fun.Map[fiber.Map](func(h database.HistoryEntry[database.HTTPRequest, database.HTTPResponse]) fiber.Map {
				b, _ := json.Marshal(h)

				var m fiber.Map
				_ = json.Unmarshal(b, &m)

				m["request_id"] = req.ID
				return m
			}, req.History.([]database.HistoryEntry[database.HTTPRequest, database.HTTPResponse])...)
		case database.SQLRequest:
			return fun.Map[fiber.Map](func(h database.HistoryEntry[database.SQLRequest, database.SQLResponse]) fiber.Map {
				b, _ := json.Marshal(h)

				var m fiber.Map
				_ = json.Unmarshal(b, &m)

				m["request_id"] = req.ID
				return m
			}, req.History.([]database.HistoryEntry[database.SQLRequest, database.SQLResponse])...)
		default:
			panic("unknown request type")
		}
	}, requests...)
	slices.SortFunc(history, func(i, j fiber.Map) int {
		return strings.Compare(j["sent_at"].(string), i["sent_at"].(string))
	})

	return fiber.Map{
		"requests": requests,
		"history":  history,
	}, nil
}

func (s *Service) HandlerCollectionUpdate(ctx context.Context, request struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}) (struct{}, error) {
	err := database.CollectionUpdate(ctx, s.DB, database.CollectionID(request.ID), request.Name)
	if err != nil {
		return struct{}{}, errors.Wrap(err, "update collection")
	}

	return struct{}{}, nil
}

func (s *Service) HandlerCollectionDelete(ctx context.Context, request struct {
	ID string `json:"id"`
}) (struct{}, error) {
	if err := database.CollectionDelete(ctx, s.DB, database.CollectionID(request.ID)); err != nil {
		return struct{}{}, errors.Wrap(err, "delete from db")
	}

	return struct{}{}, nil
}
