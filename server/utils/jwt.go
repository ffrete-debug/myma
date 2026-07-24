package utils

import (
	"sync"
	"time"

	"ark-server-commander/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var (
	tokenBlacklist     = make(map[string]time.Time)
	tokenBlacklistMutex sync.RWMutex
)

func GenerateToken(userID uint, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
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
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JWTSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
	if IsBlacklisted(tokenString) {
		return nil, jwt.ErrInvalidKey
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return config.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
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

	if time.Now().After(expiry) {
		tokenBlacklistMutex.Lock()
		delete(tokenBlacklist, tokenString)
		tokenBlacklistMutex.Unlock()
		return false
	}

	return true
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
