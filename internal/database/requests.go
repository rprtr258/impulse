package database

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	_requestSuffix = ".request.json"
	_historySuffix = ".history.jsonl"
)

func RequestGet(
	_ context.Context,
	db *DB,
	collectionID CollectionID,
	id RequestID,
) (Request, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	fs := afero.NewBasePathFs(db.fs, string(collectionID))

	var request Request
	if err := func() error {
		requestFile, err := fs.Open(string(id) + _requestSuffix)
		if err != nil {
			return errors.Wrap(err, "open request file")
		}
		defer requestFile.Close()

		b, err := io.ReadAll(requestFile)
		if err != nil {
			return errors.Wrap(err, "read request file")
		}

		if err := json.Unmarshal(b, &request); err != nil {
			return errors.Wrap(err, "parse request")
		}
		return nil
	}(); err != nil {
		return Request{}, errors.Wrap(err, "get request")
	}
	request.ID = id

	var historyPre []any
	if err := func() error {
		f, err := fs.Open(string(id) + _historySuffix)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return errors.Wrap(err, "open history file")
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		dec.DisallowUnknownFields()
		for {
			var entry any
			if err := dec.Decode(&entry); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return errors.Wrap(err, "parse history")
			}
			historyPre = append(historyPre, entry)
		}

		return nil
	}(); err != nil {
		return Request{}, errors.Wrap(err, "get history")
	}
	func() {
		b, err := json.Marshal(historyPre)
		if err != nil {
			panic(err)
		}
		switch history := request.History.(type) {
		case []HistoryEntry[HTTPRequest, HTTPResponse]:
			if err := json.Unmarshal(b, &history); err != nil {
				panic(err)
			}
			request.History = history
		case []HistoryEntry[SQLRequest, SQLResponse]:
			if err := json.Unmarshal(b, &history); err != nil {
				panic(err)
			}
			request.History = history
		default:
			panic("unknown history type")
		}
	}() // TODO: ganvnische

	return request, nil
}

type PayloadRequestCreate struct {
	Name        string
	HTTPRequest HTTPRequest
}

func RequestCreate(
	ctx context.Context,
	db *DB,
	collectionID CollectionID,
	payload PayloadRequestCreate,
) (RequestID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	fs := afero.NewBasePathFs(db.fs, string(collectionID))

	request := Request{
		RequestID(payload.Name),
		payload.HTTPRequest,
		[]HistoryEntry[HTTPRequest, HTTPResponse]{},
	}

	if err := func() error {
		requestFile, err := fs.OpenFile(string(request.ID)+_requestSuffix, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "create request %q", request.ID)
		}
		defer requestFile.Close()

		if err := json.NewEncoder(requestFile).Encode(request); err != nil {
			return errors.Wrapf(err, "write request %q", request.ID)
		}
		return nil
	}(); err != nil {
		return "", errors.Wrapf(err, "create request %q", request.ID)
	}

	if err := func() error {
		historyFile, err := fs.OpenFile(string(request.ID)+_historySuffix, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "create request %q", request.ID)
		}
		defer historyFile.Close()

		if err := json.NewEncoder(historyFile).Encode([]any{}); err != nil {
			return errors.Wrapf(err, "write request %q", request.ID)
		}
		return nil
	}(); err != nil {
		return "", errors.Wrapf(err, "create history for request %q", request.ID)
	}

	return RequestID(payload.Name), nil
}

func RequestDelete(
	ctx context.Context,
	db *DB,
	collectionID CollectionID,
	id RequestID,
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	fs := afero.NewBasePathFs(db.fs, string(collectionID))

	if err := fs.Remove(string(id) + _requestSuffix); err != nil {
		return errors.Wrapf(err, "delete request %q", id)
	}

	if err := fs.Remove(string(id) + _requestSuffix); err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "delete history %q", id)
	}

	return nil
}

func RequestUpdate(
	ctx context.Context,
	db *DB,
	collectionID CollectionID,
	id RequestID,
	kind Kind,
	newID RequestID,
	newData RequestData,
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// TODO: check kind did not change

	if id != newID {
		if err := db.fs.Rename(
			filepath.Join(string(collectionID), string(id)+_requestSuffix),
			filepath.Join(string(collectionID), string(newID)+_requestSuffix),
		); err != nil {
			return errors.Wrapf(err, "rename request %q", id)
		}
		if err := db.fs.Rename(
			filepath.Join(string(collectionID), string(id)+_historySuffix),
			filepath.Join(string(collectionID), string(newID)+_historySuffix),
		); err != nil {
			return errors.Wrapf(err, "rename history %q", id)
		}
	}

	requestFile, err := db.fs.OpenFile(filepath.Join(string(collectionID), string(id)+_requestSuffix), os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "open request %q", id)
	}
	defer requestFile.Close()

	b, err := json.Marshal(newData)
	if err != nil {
		return errors.Wrapf(err, "huita request")
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return errors.Wrapf(err, "huita 2 request")
	}
	m["kind"] = kind

	enc := json.NewEncoder(requestFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return errors.Wrapf(err, "write request %q", id)
	}

	return nil
}

func HistoryEntryCreate[I RequestData, O ResponseData](
	ctx context.Context,
	db *DB,
	collectionID CollectionID,
	id RequestID,
	item HistoryEntry[I, O],
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	fs := afero.NewBasePathFs(db.fs, string(collectionID))

	entryFilename := filepath.Join(string(id) + _historySuffix)

	entryFile, err := fs.OpenFile(entryFilename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "create history entry for request %q/%q", collectionID, id)
	}
	defer entryFile.Close()

	if _, err := entryFile.Seek(0, io.SeekEnd); err != nil {
		return errors.Wrap(err, "seek to end")
	}

	if err := json.NewEncoder(entryFile).Encode(item); err != nil {
		return errors.Wrapf(err, "write history entry for request %q/%q", collectionID, id)
	}

	return nil
}
