package service

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
)

// HandlerSend create a handler that performs call and save result to history
func (s *Service) HandlerJQ(ctx context.Context, req struct {
	JSON  string `json:"json"`
	Query string `json:"query"`
}) ([]string, error) {
	var input any
	if err := json.Unmarshal([]byte(req.JSON), &input); err != nil {
		return []string{req.JSON}, nil
	}

	result, err := func() ([]string, error) {
		query, err := gojq.Parse(req.Query)
		if err != nil {
			return nil, errors.Wrap(err, "parse query")
		}

		var result []string
		iter := query.RunWithContext(ctx, input)
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
					break
				}
				return nil, errors.Wrap(err, "run query")
			}
			b, err := json.Marshal(v)
			if err != nil {
				return nil, errors.Wrap(err, "encode result")
			}
			result = append(result, string(b))
		}
		return result, nil
	}()
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return result, nil
}
