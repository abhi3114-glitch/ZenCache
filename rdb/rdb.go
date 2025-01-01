package rdb

import (
	"encoding/gob"
	"os"
	"sync"
)

// RDB handles persistence using binary snapshots.
type RDB struct {
	mu       sync.Mutex
	filepath string
}

// NewRDB creates a new RDB instance.
func NewRDB(filepath string) *RDB {
	return &RDB{
		filepath: filepath,
	}
}

// Save writes the current data to disk.
func (r *RDB) Save(data map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.Create(r.filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(data)
}

// Load reads data from disk.
func (r *RDB) Load() (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.Open(r.filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data map[string]string
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&data)
	return data, err
}

// FilePath returns the RDB file path.
func (r *RDB) FilePath() string {
	return r.filepath
}
