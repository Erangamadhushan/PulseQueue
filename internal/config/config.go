package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func LoadConfig() (Config, error) {
	port := strings.TrimSpace(os.Getenv("PULSEQUEUE_PORT"))

	if port == "" {
		port = "8080"
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return Config{}, fmt.Errorf(
			"invalid PULSEQUEUE_PORT %q: must be between 1 and 65535",
			port,
		)
	}

	databaseURL := strings.TrimSpace(
		os.Getenv("PULSEQUEUE_DATABASE_URL"),
	)

	if databaseURL == "" {
		return Config{}, fmt.Errorf(
			"PULSEQUEUE_DATABASE_URL is required",
		)
	}

	return Config{
		Port:        port,
		DatabaseURL: databaseURL,
	}, nil
}

// os.Getenv reads an environment variable.
// strings.TrimSpace removes accidental spaces.
// strconv.Atoi converts the port string to an integer.
// Returning (Config, error) lets the caller handle invalid configuration explicitly.
// Config{} is returned when configuration is invalid, avoiding a partially configured result.
