package database

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	json2 "github.com/rprtr258/fun/exp/json"
	"github.com/spf13/afero"
)

const (
	_requestSuffix = ".request.json"
	_historySuffix = ".history.jsonl"
)

type Tree struct {
	RequestIDs []RequestID
	Dirs       map[string]Tree
}

func list(fs afero.Fs) (Tree, error) {
	infos, err := afero.ReadDir(fs, ".")
	if err != nil {
		return Tree{}, errors.Wrapf(err, "read dir")
	}

	res := Tree{
		Dirs: map[string]Tree{},
	}
	for _, info := range infos {
		if info.IsDir() {
			subdir, err := list(afero.NewBasePathFs(fs, info.Name()))
			if err != nil {
				return Tree{}, errors.Wrapf(err, "list dir %q", info.Name())
			}
			res.Dirs[info.Name()] = subdir
		} else {
			if !strings.HasSuffix(info.Name(), _requestSuffix) {
				continue
			}

			res.RequestIDs = append(res.RequestIDs, RequestID(strings.TrimSuffix(info.Name(), _requestSuffix)))
		}
	}
	return res, nil
}

func List(_ context.Context, db *DB) (Tree, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	return list(db.fs)
}

func Get(
	_ context.Context,
	db *DB,
	id RequestID,
) (Request, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var request Request
	if err := func() error {
		requestFile, err := db.fs.Open(string(id) + _requestSuffix)
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
		f, err := db.fs.Open(string(id) + _historySuffix)
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

		kind := request.Data.isRequestData()
		decoderHistory := json2.List(json2.Map4(
			func(sentAt, receivedAt time.Time, request RequestData, response ResponseData) HistoryEntry {
				return HistoryEntry{sentAt, receivedAt, request, response}
			},
			json2.Required("sent_at", json2.Time),
			json2.Required("received_at", json2.Time),
			json2.Required("request", decodersRequest[kind]),
			json2.Required("response", decodersResponse[kind]),
		))
		history, err := decoderHistory.ParseBytes(b)
		if err != nil {
			panic("unknown history type: " + err.Error())
		}
		request.History = history
	}() // TODO: ganvnische

	return request, nil
}

type PayloadRequestCreate struct {
	ID          RequestID
	RequestData RequestData
}

func Create(
	ctx context.Context,
	db *DB,
	payload PayloadRequestCreate,
) (RequestID, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	request := Request{
		RequestID(payload.ID),
		payload.RequestData,
		nil,
	}

	if err := func() error {
		requestFile, err := db.fs.OpenFile(string(request.ID)+_requestSuffix, os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			return errors.Wrap(err, "open request file")
		}
		defer requestFile.Close()

		b, err := request.MarshalJSON2()
		if err != nil {
			return errors.Wrap(err, "marshal request")
		}
		if _, err := requestFile.Write(b); err != nil {
			return errors.Wrap(err, "write request")
		}
		return nil
	}(); err != nil {
		return "", errors.Wrapf(err, "create request %q", request.ID)
	}

	if err := func() error {
		historyFile, err := db.fs.OpenFile(string(request.ID)+_historySuffix, os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			return errors.Wrap(err, "open history file")
		}
		defer historyFile.Close()

		return nil
	}(); err != nil {
		return "", errors.Wrapf(err, "create history for request %q", request.ID)
	}

	return RequestID(payload.ID), nil
}

func Delete(
	ctx context.Context,
	db *DB,
	id RequestID,
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.fs.Remove(string(id) + _requestSuffix); err != nil {
		return errors.Wrapf(err, "delete request %q", id)
	}

	if err := db.fs.Remove(string(id) + _historySuffix); err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "delete history %q", id)
	}

	return nil
}

func Update(
	ctx context.Context,
	db *DB,
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
			string(id)+_requestSuffix,
			string(newID)+_requestSuffix,
		); err != nil {
			return errors.Wrapf(err, "rename request %q", id)
		}
		if err := db.fs.Rename(
			string(id)+_historySuffix,
			string(newID)+_historySuffix,
		); err != nil {
			return errors.Wrapf(err, "rename history %q", id)
		}
	}

	requestFile, err := db.fs.OpenFile(string(id)+_requestSuffix, os.O_RDWR|os.O_TRUNC, 0o644)
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

func CreateHistoryEntry(
	ctx context.Context,
	db *DB,
	id RequestID,
	SentAt time.Time,
	ReceivedAt time.Time,
	Request RequestData,
	Response ResponseData,
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	entryFilename := string(id) + _historySuffix

	entryFile, err := db.fs.OpenFile(entryFilename, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return errors.Wrapf(err, "create history entry for request %q", id)
	}
	defer entryFile.Close()

	if _, err := entryFile.Seek(0, io.SeekEnd); err != nil {
		return errors.Wrap(err, "seek to end")
	}

	historyEntry := HistoryEntry{SentAt, ReceivedAt, Request, Response}
	if err := json.NewEncoder(entryFile).Encode(historyEntry); err != nil {
		return errors.Wrapf(err, "write history entry for request %q", id)
	}

	return nil
}
