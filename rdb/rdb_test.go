package rdb

import (
	"os"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	filepath := "test_rdb.gob"
	defer os.Remove(filepath)

	r := NewRDB(filepath)

	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	err := r.Save(data)
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	loaded, err := r.Load()
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if len(loaded) != len(data) {
		t.Errorf("Expected %d entries, got %d", len(data), len(loaded))
	}

	for k, v := range data {
		if loaded[k] != v {
			t.Errorf("Expected %s=%s, got %s", k, v, loaded[k])
		}
	}
}

func TestLoadNonexistent(t *testing.T) {
	r := NewRDB("nonexistent.gob")

	_, err := r.Load()
	if err == nil {
		t.Error("Expected error loading nonexistent file")
	}
}
