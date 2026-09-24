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

func (m *TokenMaker) generateToken(userId int64, username string, tokenType TokenType, duration time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		UserId:    userId,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

	return tokenString, nil
}