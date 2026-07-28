package utils

import (
	"errors"
	"sync"
	"time"

	"ark-server-commander/config"

	"github.com/golang-jwt/jwt/v5"
)

// Token types, carried in the "token_type" claim so a long-lived refresh token
// cannot be replayed as a short-lived access token
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// ErrInvalidTokenType is returned when a token is presented on a path that
// expects a different token type (e.g. a refresh token used as a bearer token)
var ErrInvalidTokenType = errors.New("invalid token type")

type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

var (
	tokenBlacklist       = make(map[string]time.Time)
	tokenBlacklistMutex  sync.RWMutex
	blacklistCleanupOnce sync.Once
)

func GenerateToken(userID uint, username string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JWTSecret)
}

func GenerateRefreshToken(userID uint, username string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JWTSecret)
}

// ParseToken validates a token's signature, expiry and blacklist status without
// checking its type. Request-authentication paths must use ParseAccessToken and
// the refresh endpoint must use ParseRefreshToken, otherwise a refresh token is
// accepted as a bearer token
func ParseToken(tokenString string) (*Claims, error) {
	if IsBlacklisted(tokenString) {
		return nil, jwt.ErrInvalidKey
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return config.JWTSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// ParseAccessToken parses a token and rejects it unless it is an access token.
// Tokens issued before the token_type claim existed carry an empty type and are
// still accepted, so sessions survive the upgrade
func ParseAccessToken(tokenString string) (*Claims, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeAccess && claims.TokenType != "" {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// ParseRefreshToken parses a token and rejects it unless it is a refresh token.
// Tokens issued before the token_type claim existed carry an empty type and are
// still accepted, so sessions survive the upgrade
func ParseRefreshToken(tokenString string) (*Claims, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeRefresh && claims.TokenType != "" {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

func BlacklistToken(tokenString string, expiry time.Time) {
	tokenBlacklistMutex.Lock()
	defer tokenBlacklistMutex.Unlock()
	tokenBlacklist[tokenString] = expiry
}

func IsBlacklisted(tokenString string) bool {
	tokenBlacklistMutex.RLock()
	defer tokenBlacklistMutex.RUnlock()

	expiry, exists := tokenBlacklist[tokenString]
	if !exists {
		return false
	}

	// An entry past its expiry no longer blocks anything; it is reclaimed by
	// CleanupExpiredBlacklistEntries. Deleting it here would need the write
	// lock, which a read lock cannot be upgraded to (it would deadlock)
	return !time.Now().After(expiry)
}

// StartBlacklistCleanup launches a background goroutine that sweeps expired
// blacklist entries on the given interval. Without it the blacklist grows for
// every logout and refresh and is never reclaimed. Safe to call once at startup;
// repeated calls are ignored
func StartBlacklistCleanup(interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}

	blacklistCleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for range ticker.C {
				CleanupExpiredBlacklistEntries()
			}
		}()
	})
}

func CleanupExpiredBlacklistEntries() {
	tokenBlacklistMutex.Lock()
	defer tokenBlacklistMutex.Unlock()

	now := time.Now()
	for token, expiry := range tokenBlacklist {
		if now.After(expiry) {
			delete(tokenBlacklist, token)
		}
	}
}
