package config

import (
	"os"
	"testing"
)

// setSecret points JWT_SECRET at the given value ("" means unset) and restores
// both the environment and the package globals InitConfig writes to
func setSecret(t *testing.T, secret string) {
	t.Helper()

	prevEnv, hadEnv := os.LookupEnv("JWT_SECRET")
	prevJWT, prevDB, prevPort := JWTSecret, DBPath, ServerPort

	t.Cleanup(func() {
		if hadEnv {
			os.Setenv("JWT_SECRET", prevEnv)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
		JWTSecret, DBPath, ServerPort = prevJWT, prevDB, prevPort
	})

	if secret == "" {
		os.Unsetenv("JWT_SECRET")
		return
	}
	os.Setenv("JWT_SECRET", secret)
}

func TestInitConfigRequiresSecret(t *testing.T) {
	setSecret(t, "")

	if err := InitConfig(); err == nil {
		t.Error("InitConfig should fail when JWT_SECRET is unset")
	}
	if JWTSecret != nil {
		t.Error("JWTSecret must not be set when validation fails")
	}
}

func TestInitConfigRejectsShortSecret(t *testing.T) {
	// 31 characters — one below the minimum
	setSecret(t, "kQ7vZp2XmNb4RtYu9WjLcAe6HgFdSxZ")

	if err := InitConfig(); err == nil {
		t.Error("InitConfig should fail for a secret shorter than 32 characters")
	}
}

func TestInitConfigRejectsWeakSecret(t *testing.T) {
	cases := map[string]string{
		"shipped default":   "ark-server-commander-secret-key-padded-out-to-length",
		"contains password": "correct-horse-battery-PASSWORD-stapled-on",
		"contains 123456":   "kQ7vZp2XmNb4RtYu9WjLcAe6HgFd123456",
		"contains changeme": "kQ7vZp2XmNb4RtYu9WjLcAe6HgFdchangeme",
	}

	for name, secret := range cases {
		t.Run(name, func(t *testing.T) {
			setSecret(t, secret)

			if err := InitConfig(); err == nil {
				t.Errorf("InitConfig should reject weak secret %q", secret)
			}
			if JWTSecret != nil {
				t.Error("JWTSecret must not be set when validation fails")
			}
		})
	}
}

func TestInitConfigAcceptsStrongSecret(t *testing.T) {
	const secret = "kQ7vZp2XmNb4RtYu9WjLcAe6HgFdSxZq"
	setSecret(t, secret)

	if err := InitConfig(); err != nil {
		t.Fatalf("InitConfig should accept a strong 32 character secret: %v", err)
	}
	if string(JWTSecret) != secret {
		t.Errorf("expected JWTSecret %q, got %q", secret, string(JWTSecret))
	}
}
