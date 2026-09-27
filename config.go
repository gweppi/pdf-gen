package main

import "os"

const (
	PORT          = "PORT"
	TEMPLATES_DIR = "TEMPLATES_DIR"
	GOTENBERG_URL = "GOTENBERG_URL"
	API_KEY       = "API_KEY"
)

type Config struct {
	port         string
	templatesDir string
	gotenbergUrl string
	apiKey       string
}

func LoadConfig() Config {
	return Config{
		port:         getEnv(PORT, "8080"),
		templatesDir: getEnv(TEMPLATES_DIR, "templates"),
		gotenbergUrl: getEnv(GOTENBERG_URL, "http://localhost:3000"),
		// when api key is not set, dont use it
		apiKey: getEnv(API_KEY, ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}

	return fallback
}
