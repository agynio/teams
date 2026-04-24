package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCAddress                 string
	DatabaseURL                 string
	AuthorizationServiceAddress string
	IdentityServiceAddress      string
	NotificationsServiceAddress string
}

func FromEnv() (Config, error) {
	cfg := Config{}
	cfg.GRPCAddress = os.Getenv("GRPC_ADDRESS")
	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = ":50051"
	}
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must be set")
	}
	cfg.AuthorizationServiceAddress = os.Getenv("AUTHORIZATION_SERVICE_ADDRESS")
	if cfg.AuthorizationServiceAddress == "" {
		cfg.AuthorizationServiceAddress = "authorization:50051"
	}
	cfg.IdentityServiceAddress = os.Getenv("IDENTITY_SERVICE_ADDRESS")
	if cfg.IdentityServiceAddress == "" {
		cfg.IdentityServiceAddress = "identity:50051"
	}
	cfg.NotificationsServiceAddress = os.Getenv("NOTIFICATIONS_SERVICE_ADDRESS")
	if cfg.NotificationsServiceAddress == "" {
		cfg.NotificationsServiceAddress = "notifications:50051"
	}
	return cfg, nil
}
