package backend

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddress      string
	DatabasePath     string
	StaticPath       string
	Environment      string
	AllowedWSOrigins []string
}

func LoadConfig() Config {
	config := Config{
		HTTPAddress:  envOrDefault("FORUM_HTTP_ADDRESS", ":8080"),
		DatabasePath: envOrDefault("FORUM_DATABASE_PATH", "database/forum.db"),
		StaticPath:   envOrDefault("FORUM_STATIC_PATH", "static"),
		Environment:  envOrDefault("FORUM_ENV", "development"),
	}

	origins := os.Getenv("FORUM_WS_ORIGINS")
	if origins == "" {
		config.AllowedWSOrigins = []string{
			"http://localhost:8080",
			"http://127.0.0.1:8080",
		}
	} else {
		for _, origin := range strings.Split(origins, ",") {
			if origin = strings.TrimSpace(origin); origin != "" {
				config.AllowedWSOrigins = append(config.AllowedWSOrigins, origin)
			}
		}
	}

	return config
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func (c Config) SecureCookies() bool {
	return c.Environment == "production"
}
