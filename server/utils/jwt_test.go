package utils

import (
	"errors"
	"testing"
	"time"

	"ark-server-commander/config"

	"github.com/golang-jwt/jwt/v5"
)

// setupJWT plants a signing key: config.InitConfig never runs in unit tests, so
// config.JWTSecret would otherwise be nil
func setupJWT(t *testing.T) {
	t.Helper()
	config.JWTSecret = []byte("unit-test-jwt-signing-key-0123456789")
}

func TestParseTokenRejectsBlacklistedToken(t *testing.T) {
	setupJWT(t)

	token, err := GenerateToken(1, "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if _, err := ParseToken(token); err != nil {
		t.Fatalf("fresh token should parse: %v", err)
	}

	BlacklistToken(token, time.Now().Add(time.Hour))

	if _, err := ParseToken(token); err == nil {
		t.Error("blacklisted token should be rejected")
	}
	if _, err := ParseAccessToken(token); err == nil {
		t.Error("blacklisted token should be rejected by ParseAccessToken")
	}
}

// TestIsBlacklistedExpiredEntryDoesNotDeadlock is the regression test for the
// read lock that was upgraded to a write lock in order to evict an expired
// entry — sync.RWMutex cannot be upgraded, so that path hung every request
func TestIsBlacklistedExpiredEntryDoesNotDeadlock(t *testing.T) {
	setupJWT(t)

	token := "expired-blacklist-entry"
	BlacklistToken(token, time.Now().Add(-time.Hour))

	done := make(chan bool, 1)
	go func() {
		done <- IsBlacklisted(token)
	}()

	select {
	case blacklisted := <-done:
		if blacklisted {
			t.Error("an expired blacklist entry should no longer block the token")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("IsBlacklisted deadlocked on an expired blacklist entry")
	}

	// The entry is only reclaimed by the sweeper, which takes the write lock
	CleanupExpiredBlacklistEntries()
	if IsBlacklisted(token) {
		t.Error("token should not be blacklisted after cleanup")
	}
}

func TestParseAccessTokenRejectsRefreshToken(t *testing.T) {
	setupJWT(t)

	refresh, err := GenerateRefreshToken(2, "admin")
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if _, err := ParseAccessToken(refresh); !errors.Is(err, ErrInvalidTokenType) {
		t.Errorf("refresh token must not be usable as an access token, got err=%v", err)
	}
	if claims, err := ParseRefreshToken(refresh); err != nil {
		t.Errorf("refresh token should parse on the refresh path: %v", err)
	} else if claims.TokenType != TokenTypeRefresh {
		t.Errorf("expected token_type %q, got %q", TokenTypeRefresh, claims.TokenType)
	}
}

func TestParseRefreshTokenRejectsAccessToken(t *testing.T) {
	setupJWT(t)

	access, err := GenerateToken(3, "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if _, err := ParseRefreshToken(access); !errors.Is(err, ErrInvalidTokenType) {
		t.Errorf("access token must not be usable on the refresh path, got err=%v", err)
	}
	if claims, err := ParseAccessToken(access); err != nil {
		t.Errorf("access token should parse on the access path: %v", err)
	} else if claims.TokenType != TokenTypeAccess {
		t.Errorf("expected token_type %q, got %q", TokenTypeAccess, claims.TokenType)
	}
}

func TestParseTokenRejectsUnexpectedSigningMethod(t *testing.T) {
	setupJWT(t)

	claims := &Claims{
		UserID:    4,
		Username:  "admin",
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Same secret, different HMAC variant: only HS256 is allowed
	hs512, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(config.JWTSecret)
	if err != nil {
		t.Fatalf("HS512 signing failed: %v", err)
	}
	if _, err := ParseToken(hs512); err == nil {
		t.Error("token signed with HS512 should be rejected")
	}

	// The classic alg=none downgrade
	none, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("none signing failed: %v", err)
	}
	if _, err := ParseToken(none); err == nil {
		t.Error("token signed with alg=none should be rejected")
	}
}
