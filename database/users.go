package database

import (
	"errors"
	"fmt"
)

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func (db *DB) CreateUser(body string) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, fmt.Errorf("failed to load database: %s", err)
	}

	id := len(data.Users) + 1
	user := User{Email: body, ID: id}
	data.Users[id] = user

	db.writeDB(data)

	return user, nil
}

func (db *DB) GetUsers() ([]User, error) {
	data, err := db.loadDB()
	if err != nil {
		return nil, errors.New("failed to load database")
	}

	users := make([]User, 0, len(data.Users))

	for _, user := range data.Users {
		users = append(users, user)
	}

	return users, nil
}

func (db *DB) GetUser(id int) (*User, error) {
	users, err := db.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %s", err)
	}

	for _, user := range users {
		if user.ID == id {
			return &user, nil
		}
	}

	return nil, nil
}
