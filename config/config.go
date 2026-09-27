package config

import "os"

const (
	PORT          = "PORT"
	GOTENBERG_URL = "GOTENBERG_URL"
	API_KEY       = "API_KEY"
	HOSTNAME      = "HOSTNAME"
)

type Config struct {
	Port         string
	TemplatesDir string
	OutputDir    string
	GotenbergUrl string
	ApiKey       string
	Hostname     string
}

func Load() Config {
	return Config{
		Port:         getEnv(PORT, "8080"),
		TemplatesDir: "templates",
		OutputDir:    "output",
		GotenbergUrl: getEnv(GOTENBERG_URL, "http://localhost:3000"),
		// when api key is not set, dont use it
		ApiKey:   getEnv(API_KEY, ""),
		Hostname: getEnv(HOSTNAME, "http://localhost:8080/"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}

	return fallback
}
