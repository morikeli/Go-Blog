package utils

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
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
	tokenId := uuid.New().String()

	claims := Claims{
		UserId:    userId,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenId,
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

		// Why it's there: This checks if the algorithm used to sign the token belongs to the HMAC family (symmetric key signing).
		// Security Purpose: Prevents algorithm confusion attacks (e.g., an attacker changing the header to use an asymmetric method like RSA,
		// causing the parser to treat your public key as a secret HMAC key).
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Invalid signing method!")
		}

		// Specifically require HS256.

		// Why it's there: Even within the HMAC family (which includes HS256, HS384, HS512), this ensures the token was
		// specifically signed using HS256, matching the exact method used when tokens are created in
		// [generateToken] (jwt.NewWithClaims(jwt.SigningMethodHS256, claims)).
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("Unexpected signing algorithm!")
		}

		return m.secretKey, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc, jwt.WithIssuer(m.issuer))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("Invalid token: token has expired!")
		}

		if errors.Is(err, jwt.ErrTokenInvalidIssuer) {
			return nil, errors.New("Invalid token: invalid issuer!")
		}

		return nil, errors.New("Invalid token: token may be malformed or tampered with!")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid token: token claims are invalid!")
	}

	return claims, nil
}

func (m *TokenMaker) VerifyAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, errors.New("Invalid access token!")
	}

	return claims, nil
}

func (m *TokenMaker) VerifyRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, errors.New("Invalid refresh token!")
	}

	return claims, nil
}

func RevokeRefreshToken(ctx context.Context, rdb *redis.Client, claims *Claims) error {
	if claims == nil || claims.ID == "" {
		return nil
	}

	if claims.ExpiresAt == nil {
		return nil
	}

	remainingDuration := time.Until(claims.ExpiresAt.Time)

	if remainingDuration <= 0 {
		return nil
	}

	blacklistKey := "blacklist:" + claims.ID

	return rdb.Set(ctx, blacklistKey, "revoked", remainingDuration).Err()
}
