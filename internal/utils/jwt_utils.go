package utils

import (
	"errors"
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
	UserId    int64     `json:"user_id"`
	Username  string    `json:"username"`
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
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *TokenMaker) GenerateAccessToken(userId int64, username string, duration time.Duration) (string, error) {
	return m.generateToken(userId, username, AccessToken, duration)
}

func (m *TokenMaker) GenerateRefreshToken(userId int64, username string, duration time.Duration) (string, error) {
	return m.generateToken(userId, username, RefreshToken, duration)
}

func (m *TokenMaker) VerifyToken(tokenString string) (*Claims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Ensure signing method matches expectations
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Invalid signing method!")
		}
		return m.secretKey, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("Invalid token: token has expired!")
		}
		return nil, errors.New("Invalid token: token may be malformed or tampered with!")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid token: token claims are invalid!")
	}

	return claims, nil
}
