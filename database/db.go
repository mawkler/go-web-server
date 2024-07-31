package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type DB struct {
	mux       *sync.RWMutex
	path      string
	idCounter int
}

type DBStructure struct {
	Chirps      map[int]Chirp    `json:"chirps"`
	Users       map[int]FullUser `json:"users"`
	Revocations []string         `json:"revocations"`
}

func New(path string) *DB {
	return &DB{mux: &sync.RWMutex{}, path: path}
}

func (db *DB) ensureDB() error {
	f, err := os.Create(db.path)
	if err != nil {
		return errors.New("failed to create database file")
	}

	_, err = io.WriteString(f, `{"chirps": {}, "users": {}, "revocations": []}`)
	if err != nil {
		return fmt.Errorf("failed to write to file: %s", err)
	}

	defer f.Close()

	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	file, err := os.ReadFile(db.path)
	if err != nil {
		db.ensureDB()
		file, err = os.ReadFile(db.path)
		if err != nil {
			return DBStructure{}, fmt.Errorf("could not create database file: %s", err)
		}
	}

	data := DBStructure{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		return DBStructure{}, fmt.Errorf("could not marshal database: %s", err)
	}

	return data, nil
}

func (db *DB) writeDB(data DBStructure) error {
	marshalledData, err := json.Marshal(data)
	if err != nil {
		return errors.New("failed to marshall data")
	}

	return os.WriteFile(db.path, marshalledData, 0666)
}
