package config

import (
	"github.com/AbsaOSS/gopkg/env"
	"github.com/pkg/errors"
)

const (
	idpEndpointKey     = "IDP_ENDPOINT"
	awsRegionKey       = "AWS_REGION"
	sessionDurationKey = "SESSION_DURATION"
)

// Config structure
type Config struct {
	IDPEndpoint     string
	AWSRegion       string
	SessionDuration string
}

// GetConfig returns operator config structure
func GetConfig() (*Config, error) {
	idpEndpoint := env.GetEnvAsStringOrFallback(idpEndpointKey, "")
	awsRegion := env.GetEnvAsStringOrFallback(awsRegionKey, "")
	sessionDuration := env.GetEnvAsStringOrFallback(sessionDurationKey, "1h")
	if idpEndpoint == "" {
		return nil, errors.New("IDPEndpoint can't be empty")
	}
	return &Config{
		IDPEndpoint:     idpEndpoint,
		SessionDuration: sessionDuration,
		AWSRegion:       awsRegion,
	}, nil
}
