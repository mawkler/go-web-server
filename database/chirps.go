package database

import (
	"errors"
	"fmt"
)

type Chirp struct {
	Body     string `json:"body"`
	ID       int    `json:"id"`
	AuthorID int    `json:"author_id"`
}

func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, fmt.Errorf("failed to load database: %s", err)
	}

	id := len(data.Chirps) + 1
	chirp := Chirp{Body: body, ID: id, AuthorID: authorID}
	data.Chirps[id] = chirp

	db.writeDB(data)

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return nil, errors.New("failed to load database")
	}

	chirps := make([]Chirp, 0, len(data.Chirps))

	for _, chirp := range data.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) GetChirp(id int) (*Chirp, error) {
	chirps, err := db.GetChirps()
	if err != nil {
		return nil, fmt.Errorf("failed to get chirp: %s", err)
	}

	for _, chirp := range chirps {
		if chirp.ID == id {
			return &chirp, nil
		}
	}

	return nil, nil
}

func (db *DB) GetChirpsByAuthor(authorID int) ([]Chirp, error) {
	chirps, err := db.GetChirps()
	authorChirps := []Chirp{}

	if err != nil {
		return nil, fmt.Errorf("failed to get author %d's chirps: %s", authorID, err)
	}

	for _, chirp := range chirps {
		if chirp.AuthorID == authorID {
			authorChirps = append(authorChirps, chirp)
		}
	}

	return authorChirps, nil
}

func (db *DB) DeleteChirp(id int) error {
	data, err := db.loadDB()
	if err != nil {
		return errors.New("failed to load database")
	}

	delete(data.Chirps, id)

	if err := db.writeDB(data); err != nil {
		return fmt.Errorf("failed to delete chirp %d: %s", id, err)
	}

	return nil
}
