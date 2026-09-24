package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)
type TokenMaker struct {
	secretKey []byte
	issuer    string
}

type Claims struct {
	UserId int64 `json:"user_id"`
	Username string `json:"username"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

func NewTokenMaker(secretKey string, issuer string) *TokenMaker {
	return &TokenMaker{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

func (m *TokenMaker) GenerateToken(userId int64, username string) (string, error) {
	claims := Claims{
		UserId:    userId,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}