package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host           string
	Port           string
	DatabaseURL    string
	APIToken       string
	FrontendOrigin string
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

	host := strings.TrimSpace(os.Getenv("PULSEQUEUE_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}

	apiToken := strings.TrimSpace(os.Getenv("PULSEQUEUE_API_TOKEN"))
	if !isLoopbackHost(host) && apiToken == "" {
		return Config{}, fmt.Errorf(
			"PULSEQUEUE_API_TOKEN is required when PULSEQUEUE_HOST is not loopback",
		)
	}

	frontendOrigin := strings.TrimSpace(os.Getenv("PULSEQUEUE_FRONTEND_ORIGIN"))
	if frontendOrigin != "" {
		origin, err := url.Parse(frontendOrigin)
		if err != nil ||
			(origin.Scheme != "http" && origin.Scheme != "https") ||
			origin.Host == "" ||
			origin.Path != "" ||
			origin.RawQuery != "" ||
			origin.Fragment != "" ||
			origin.User != nil {
			return Config{}, fmt.Errorf(
				"invalid PULSEQUEUE_FRONTEND_ORIGIN %q: must be an HTTP(S) origin",
				frontendOrigin,
			)
		}
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
		Host:           host,
		Port:           port,
		DatabaseURL:    databaseURL,
		APIToken:       apiToken,
		FrontendOrigin: frontendOrigin,
	}, nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// os.Getenv reads an environment variable.
// strings.TrimSpace removes accidental spaces.
// strconv.Atoi converts the port string to an integer.
// Returning (Config, error) lets the caller handle invalid configuration explicitly.
// Config{} is returned when configuration is invalid, avoiding a partially configured result.
