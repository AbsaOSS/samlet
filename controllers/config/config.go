package config

import (
	// this is a copy of https://github.com/AbsaOSS/k8gb/blob/master/controllers/internal/env/env.go
	// https://github.com/AbsaOSS/k8gb/issues/185
	"github.com/bison-cloud-platform/samlet/controllers/internal/env"
	"github.com/pkg/errors"
)

const (
	idpEndpointKey     = "IDP_ENDPOINT"
	sessionDurationKey = "SESSION_DURATION"
)

// Config structure
type Config struct {
	IDPEndpoint     string
	SessionDuration int
}

// GetConfig returns operator config structure
func GetConfig() (*Config, error) {
	idpEndpoint := env.GetEnvAsStringOrFallback(idpEndpointKey, "")
	sessionDuration, err := env.GetEnvAsIntOrFallback(sessionDurationKey, 3600)
	if err != nil {
		return nil, err
	}
	if idpEndpoint == "" {
		return nil, errors.New("IDPEndpoint can't be empty")
	}
	return &Config{
		IDPEndpoint:     idpEndpoint,
		SessionDuration: sessionDuration,
	}, nil
}
