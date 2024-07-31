package database

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// TODO: move token stuff to a separate file
type RefreshToken struct {
	ExpiresAt time.Time `json:"expires_at"`
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
}

type User struct {
	Email string `json:"email"`
	ID    int    `json:"id"`
}

type FullUser struct {
	Password string `json:"password"`
	User
}

func (u *FullUser) toUser() *User {
	return &User{Email: u.Email, ID: u.ID}
}

func hashPassword(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %s", err)
	}

	return string(passwordHash), nil
}

func (db *DB) CreateUser(email, password string) (User, error) {
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
		User:     User{Email: email, ID: id},
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

	updatedUser := FullUser{
		User:     User{Email: email, ID: id},
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

func (db *DB) SaveRefreshToken(refreshToken string, userID int, expiresIn time.Duration) error {
	data, err := db.loadDB()
	if err != nil {
		return fmt.Errorf("failed to load database: %s", err)
	}

	expiresAt := time.Now().Add(time.Second * expiresIn)
	token := RefreshToken{
		Token:     refreshToken,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	data.RefreshTokens[refreshToken] = token
	if err := db.writeDB(data); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %s", err)
	}

	return nil
}

func (db *DB) RevokeRefreshToken(refreshToken string) error {
	data, err := db.loadDB()
	if err != nil {
		return fmt.Errorf("failed to load database: %s", err)
	}

	delete(data.RefreshTokens, refreshToken)
	if err := db.writeDB(data); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %s", err)
	}

	return nil
}

func (db *DB) GetRefreshToken(refreshToken string) (*RefreshToken, error) {
	data, err := db.loadDB()
	if err != nil {
		return nil, errors.New("failed to load database")
	}

	token, exists := data.RefreshTokens[refreshToken]
	if exists {
		return &token, nil
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
