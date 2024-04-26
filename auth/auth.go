package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func createJwt(userID int, issuer, jwtSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   fmt.Sprint(userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %s", err)
	}

	return signedToken, nil
}

func CreateAccessToken(userID int, jwtSecret string) (string, error) {
	return createJwt(userID, "chirpy-access", jwtSecret, 1*time.Hour)
}

func CreateRefreshToken(userID int, jwtSecret string) (string, error) {
	return createJwt(userID, "chirpy-refresh", jwtSecret, 60*24*time.Hour)
}

func Authorize(tokenString string, jwtSecret string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %s", err)
	}
	return token, nil
}
