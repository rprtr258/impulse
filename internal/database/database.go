package database

import (
	"sync"

	"github.com/spf13/afero"
)

type DB struct {
	fs afero.Fs
	mu sync.Mutex
}

func New(fs afero.Fs) *DB {
	return &DB{
		fs: fs,
	}
}

func (db *DB) Close() error {
	return nil
}
