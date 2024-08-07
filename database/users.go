package database

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email       string `json:"email"`
	ID          int    `json:"id"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

type FullUser struct {
	Password string `json:"password"`
	User
}

func (u *FullUser) toUser() *User {
	return &User{Email: u.Email, ID: u.ID, IsChirpyRed: u.IsChirpyRed}
}

func hashPassword(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %s", err)
	}

	return string(passwordHash), nil
}

func (db *DB) CreateUser(email, password string, isChirpyRed bool) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, fmt.Errorf("failed to load database: %s", err)
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}

	id := len(data.Users) + 1
	user := FullUser{
		User:     User{Email: email, ID: id, IsChirpyRed: isChirpyRed},
		Password: string(passwordHash),
	}
	data.Users[id] = user

	err = db.writeDB(data)
	if err != nil {
		return User{}, fmt.Errorf("failed to write user to database: %s", err)
	}

	return *user.toUser(), nil
}

func (db *DB) UpdateUser(id int, email, password string) (*User, error) {
	user, err := db.GetUser(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %s", email, err)
	}

	if user == nil {
		return nil, nil
	}

	data, err := db.loadDB()
	if err != nil {
		return nil, fmt.Errorf("failed to load database: %s", err)
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user.Email = email
	updatedUser := FullUser{
		User:     *user,
		Password: passwordHash,
	}
	data.Users[id] = updatedUser

	err = db.writeDB(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write user to database: %s", err)
	}

	return updatedUser.toUser(), nil
}

func (db *DB) GetUsers() ([]User, error) {
	data, err := db.loadDB()
	if err != nil {
		return nil, errors.New("failed to load database")
	}

	users := make([]User, 0, len(data.Users))

	for _, user := range data.Users {
		users = append(users, *user.toUser())
	}

	return users, nil
}

func (db *DB) getUsersWithPasswords() ([]FullUser, error) {
	data, err := db.loadDB()
	if err != nil {
		return nil, errors.New("failed to load database")
	}

	users := make([]FullUser, 0, len(data.Users))

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

func (db *DB) getUserByEmail(email string) (*FullUser, error) {
	users, err := db.getUsersWithPasswords()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %s", err)
	}

	for _, user := range users {
		if user.Email == email {
			return &user, nil
		}
	}

	return nil, nil
}

func (db *DB) UpgradeUser(id int) error {
	data, err := db.loadDB()
	if err != nil {
		return fmt.Errorf("failed to upgrade user: %s", err)
	}

	user := data.Users[id]
	user.IsChirpyRed = true
	data.Users[id] = user

	db.writeDB(data)
	return nil
}
