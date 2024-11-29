package service

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/impulse-http/local-backend/internal/database"
)

func (s *Service) sendHTTP(ctx context.Context, req database.HTTPRequest) (database.HTTPResponse, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		req.Method,
		req.URL,
		strings.NewReader(req.Body),
	)
	if err != nil {
		return database.HTTPResponse{}, errors.Wrap(err, "create request")
	}
	request.Header = fromKV(req.Headers)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return database.HTTPResponse{}, errors.Wrap(err, "perform request")
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(response.Body); err != nil {
		return database.HTTPResponse{}, err
	}

	return database.HTTPResponse{
		response.StatusCode,
		buf.String(),
		toKV(response.Header),
	}, nil
}
