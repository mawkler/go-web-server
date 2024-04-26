package database

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func (db DB) Login(email, password string) (*User, error) {
	user, err := db.getUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("login failed: %s", err)
	}

	if user == nil {
		log.Printf("login failed, user %s does not exist", email)
		return nil, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password: %s", err)
	}

	return &user.User, nil
}
