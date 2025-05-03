package database

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	json2 "github.com/rprtr258/fun/exp/json"
	"github.com/spf13/afero"
	"go.nhat.io/aferocopy/v2"
)

const (
	_requestSuffix = ".request.json"
	_historySuffix = ".history.jsonl"
)

type Tree struct {
	RequestIDs []RequestID
	Dirs       map[string]Tree
}

func list(fs afero.Fs, prefix string) (Tree, error) {
	infos, err := afero.ReadDir(fs, prefix)
	if err != nil {
		return Tree{}, errors.Wrapf(err, "read dir")
	}

	res := Tree{
		[]RequestID{},
		map[string]Tree{},
	}
	for _, info := range infos {
		if info.IsDir() {
			dir := prefix + info.Name()
			subdir, err := list(fs, dir+"/")
			if err != nil {
				return Tree{}, errors.Wrapf(err, "list dir %q", info.Name())
			}
			res.Dirs[dir] = subdir
		} else {
			baseRequestID, isRequest := strings.CutSuffix(info.Name(), _requestSuffix)
			if !isRequest {
				continue
			}

			res.RequestIDs = append(res.RequestIDs, RequestID(prefix+baseRequestID))
		}
	}
	return res, nil
}

func List(_ context.Context, db *DB) (Tree, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	return list(db.fs, "")
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
		if string(b) == "null" {
			return
		}

		history, err := json2.List(DecodeHistory(request.Data)).ParseBytes(b)
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

func Duplicate(
	ctx context.Context,
	db *DB,
	id string,
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, err := db.fs.Stat(string(id) + _requestSuffix); err == nil { // it is request file, duplicate
		n := 1
		for {
			if _, err := db.fs.Stat(string(id) + " (" + strconv.Itoa(n) + ")" + _requestSuffix); err == nil {
				n++
			} else {
				break
			}
		}

		if err := aferocopy.Copy(string(id)+_requestSuffix, string(id)+" ("+strconv.Itoa(n)+")"+_requestSuffix, aferocopy.Options{
			SrcFs:  db.fs,
			DestFs: db.fs,
			Sync:   true,
		}); err != nil {
			return errors.Wrapf(err, "duplicate request %q", id)
		}
	} else if stat, err := db.fs.Stat(string(id)); err == nil && stat.IsDir() { // it is dir, duplicate
		n := 1
		for {
			if _, err := db.fs.Stat(string(id) + " (" + strconv.Itoa(n) + ")"); err == nil {
				n++
			} else {
				break
			}
		}

		if err := aferocopy.Copy(string(id), string(id)+" ("+strconv.Itoa(n)+")", aferocopy.Options{
			SrcFs:  db.fs,
			DestFs: db.fs,
			Sync:   true,
		}); err != nil {
			return errors.Wrapf(err, "duplicate dir %q", id)
		}
	} else {
		return errors.Errorf("unknown request/dir %q", id)
	}

	return nil
}

func Delete(
	ctx context.Context,
	db *DB,
	id RequestID,
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, err := db.fs.Stat(string(id) + _requestSuffix); err == nil { // it is request file, remove
		if err := db.fs.Remove(string(id) + _requestSuffix); err != nil {
			return errors.Wrapf(err, "delete request %q", id)
		}

		if err := db.fs.Remove(string(id) + _historySuffix); err != nil && !os.IsNotExist(err) {
			return errors.Wrapf(err, "delete history %q", id)
		}
	} else if stat, err := db.fs.Stat(string(id)); err == nil && stat.IsDir() { // it is dir, remove
		if err := db.fs.RemoveAll(string(id)); err != nil {
			return errors.Wrapf(err, "delete request %q", id)
		}
	} else {
		return errors.Errorf("unknown request/dir %q", id)
	}

	return nil
}

func Rename(
	ctx context.Context,
	db *DB,
	id, newID RequestID,
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// check target files in case they already exist
	{
		if _, err := db.fs.Stat(string(newID) + _requestSuffix); err == nil || !os.IsNotExist(err) {
			if err != nil {
				return errors.Wrapf(err, "check target request file %q", newID)
			}
			return errors.Errorf("target request file %q already exist, delete it first", newID)
		}
		if _, err := db.fs.Stat(string(newID) + _historySuffix); err == nil || !os.IsNotExist(err) {
			if err != nil {
				return errors.Wrapf(err, "check target history file %q", newID)
			}
			return errors.Errorf("target history file %q already exist, delete it first", newID)
		}
	}

	{
		if dir := filepath.Dir(string(newID)); dir != "." {
			if err := db.fs.MkdirAll(dir, 0o755); err != nil {
				return errors.Wrapf(err, "create dir %q", dir)
			}
		}

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
			if !os.IsNotExist(err) { // NOTE: history might not exist
				return errors.Wrapf(err, "rename history %q", id)
			}
		}
	}

	return nil
}

func Update(
	ctx context.Context,
	db *DB,
	id RequestID,
	kind Kind,
	newData RequestData,
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if kind != newData.Kind() {
		return errors.Errorf("kind mismatch: %q != %q", kind, newData.Kind())
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
