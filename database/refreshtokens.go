package database

import (
	"errors"
	"fmt"
	"time"
)

type RefreshToken struct {
	ExpiresAt time.Time `json:"expires_at"`
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
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

func (db *DB) DeleteRefreshToken(refreshToken string) error {
	data, err := db.loadDB()
	if err != nil {
		return fmt.Errorf("failed to load database: %s", err)
	}

	delete(data.RefreshTokens, refreshToken)
	if err := db.writeDB(data); err != nil {
		return fmt.Errorf("failed to delete refresh token: %s", err)
	}

	return nil
}
