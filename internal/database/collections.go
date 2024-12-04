package database

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/rprtr258/fun"
	"github.com/spf13/afero"
)

func CollectionCreate(ctx context.Context, db *DB, name string) (Collection, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.fs.Mkdir(name, os.ModePerm); err != nil && !os.IsExist(err) {
		return Collection{}, errors.Wrap(err, "create collection")
	}

	return Collection{
		CollectionID(name),
		name,
		nil,
	}, nil
}

func CollectionDelete(ctx context.Context, db *DB, id CollectionID) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.fs.RemoveAll(string(id)); err != nil {
		return errors.Wrap(err, "delete collection")
	}

	return nil
}

func CollectionList(ctx context.Context, db *DB) ([]Collection, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	collectionDirs, err := afero.ReadDir(db.fs, ".")
	if err != nil {
		return nil, errors.Wrap(err, "list collections")
	}

	collections := make([]Collection, 0, len(collectionDirs))
	for _, stat := range collectionDirs {
		if !stat.IsDir() {
			continue
		}

		requestFiles, err := afero.Glob(afero.NewBasePathFs(db.fs, stat.Name()), "*"+_requestSuffix)
		if err != nil {
			return nil, errors.Wrapf(err, "list requests in collection %q", stat.Name())
		}

		collections = append(collections, Collection{
			CollectionID(stat.Name()),
			stat.Name(),
			fun.Map[RequestID](
				func(name string) RequestID { return RequestID(strings.TrimSuffix(name, _requestSuffix)) },
				requestFiles...),
		})
	}
	sort.Slice(collections, func(i, j int) bool {
		return collections[i].ID < collections[j].ID
	})
	return collections, nil
}

func CollectionUpdate(ctx context.Context, db *DB, id CollectionID, name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.fs.Rename(string(id), name); err != nil {
		return errors.Wrap(err, "rename collection")
	}

	return nil
}

func CollectionGet(ctx context.Context, db *DB, id CollectionID) (Collection, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	requestFiles, err := afero.Glob(afero.NewBasePathFs(db.fs, string(id)), "*"+_requestSuffix)
	if err != nil {
		return Collection{}, errors.Wrapf(err, "list requests in collection %q", id)
	}

	return Collection{
		id,
		string(id),
		fun.Map[RequestID](
			func(name string) RequestID { return RequestID(strings.TrimSuffix(name, _requestSuffix)) },
			requestFiles...),
	}, nil
}
