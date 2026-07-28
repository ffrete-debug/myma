package config

import (
	"fmt"
	"os"
	"strings"
)

var (
	JWTSecret  []byte
	DBPath     = "ark_server.db"
	ServerPort = "8080"

	// RCONHost is the address used to reach a game server's RCON port.
	//
	// Game servers run as separate containers and publish their RCON port on
	// the Docker host. This process therefore cannot use "localhost": inside a
	// container that resolves to this container, not the game server, so every
	// dial fails. When containerised we go out via the Docker host gateway to
	// the published port; on a bare-metal run the loopback address is correct.
	// Override with RCON_HOST when the game servers live somewhere else.
	RCONHost = "127.0.0.1"

	// SteamAPIKey enables Steam Workshop *search*. Looking a mod up by its
	// Workshop ID needs no key; only the search endpoint does. Empty means the
	// mod browser offers add-by-ID but reports search as unconfigured.
	SteamAPIKey = ""

	// S3-compatible object storage for automated off-host backups. Empty
	// values mean cloud upload is simply unavailable; local backups still work.
	S3Endpoint  = ""
	S3Region    = "us-east-1"
	S3Bucket    = ""
	S3AccessKey = ""
	S3SecretKey = ""
	S3Prefix    = ""
	// S3PathStyle is required by MinIO and most self-hosted gateways, which do
	// not implement virtual-host bucket addressing.
	S3PathStyle = false
)

// dockerHostGateway is the alias Docker resolves to the host when the container
// is started with `--add-host host.docker.internal:host-gateway` (set for this
// service in docker-compose.yml).
const dockerHostGateway = "host.docker.internal"

// runningInContainer reports whether this process is inside a container.
// /.dockerenv is created by the Docker daemon in every container it starts.
func runningInContainer() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

// Weak secret blacklist
var weakSecrets = []string{
	"ark-server-commander-secret-key",
	"secret",
	"password",
	"123456",
	"default",
	"changeme",
	"test",
}

func InitConfig() error {
	// JWT secret must be read from environment variable
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}

	// Validate secret length
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long (current: %d)", len(secret))
	}

	// Check for weak secrets
	secretLower := strings.ToLower(secret)
	for _, weak := range weakSecrets {
		if strings.Contains(secretLower, weak) {
			return fmt.Errorf("JWT_SECRET contains weak/common password pattern: '%s'", weak)
		}
	}

	JWTSecret = []byte(secret)

	// Read other configuration
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		DBPath = dbPath
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		ServerPort = port
	}

	SteamAPIKey = os.Getenv("STEAM_API_KEY")

	S3Endpoint = os.Getenv("S3_ENDPOINT")
	S3Bucket = os.Getenv("S3_BUCKET")
	S3AccessKey = os.Getenv("S3_ACCESS_KEY")
	S3SecretKey = os.Getenv("S3_SECRET_KEY")
	S3Prefix = os.Getenv("S3_PREFIX")
	S3PathStyle = os.Getenv("S3_PATH_STYLE") == "true"
	if region := os.Getenv("S3_REGION"); region != "" {
		S3Region = region
	}

	if host := os.Getenv("RCON_HOST"); host != "" {
		RCONHost = host
	} else if runningInContainer() {
		RCONHost = dockerHostGateway
	}

	return nil
}
