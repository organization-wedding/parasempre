package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	envDBHost     = "DB_HOST"
	envDBPort     = "DB_PORT"
	envDBUser     = "DB_USER"
	envDBPassword = "DB_PASSWORD"
	envDBName     = "DB_NAME"
	envDBSSLMode  = "DB_SSLMODE"

	envCORSOrigin = "CORS_ORIGIN"
	envAppEnv     = "APP_ENV"

	envGroomPhone = "GROOM_PHONE"
	envGroomURACF = "GROOM_URACF"
	envBridePhone = "BRIDE_PHONE"
	envBrideURACF = "BRIDE_URACF"

	envJWTSecret = "JWT_SECRET"
	envJWTExpiry = "JWT_EXPIRY"

	envEvoAPIURL      = "EVO_API_URL"
	envEvoAPIKey      = "EVO_API_KEY"
	envEvoAPIInstance = "EVO_API_INSTANCE"
)

const (
	defaultCORSOrigin = "http://localhost:3000"
	defaultAppEnv     = "test"
	defaultJWTExpiry  = "3h"
	defaultDBSSLMode  = "require"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type CoupleUserConfig struct {
	Phone string
	URACF string
}

type CoupleConfig struct {
	Groom CoupleUserConfig
	Bride CoupleUserConfig
}

type Config struct {
	DB             DBConfig
	CORSOrigin     string
	AppEnv         string
	Couple         CoupleConfig
	JWTSecret      string
	JWTExpiry      string
	EvoAPIURL      string
	EvoAPIKey      string
	EvoAPIInstance string
}

type envField struct {
	name  string
	value string
}

func Load() (Config, error) {
	cfg := Config{
		DB: DBConfig{
			Host:     getEnv(envDBHost),
			Port:     getEnv(envDBPort),
			User:     getEnv(envDBUser),
			Password: getEnv(envDBPassword),
			Name:     getEnv(envDBName),
			SSLMode:  getEnvOrDefault(envDBSSLMode, defaultDBSSLMode),
		},
		CORSOrigin: getEnvOrDefault(envCORSOrigin, defaultCORSOrigin),
		AppEnv:     getEnvOrDefault(envAppEnv, defaultAppEnv),
		Couple: CoupleConfig{
			Groom: CoupleUserConfig{
				Phone: getEnv(envGroomPhone),
				URACF: getEnv(envGroomURACF),
			},
			Bride: CoupleUserConfig{
				Phone: getEnv(envBridePhone),
				URACF: getEnv(envBrideURACF),
			},
		},
		JWTSecret:      getEnv(envJWTSecret),
		JWTExpiry:      getEnvOrDefault(envJWTExpiry, defaultJWTExpiry),
		EvoAPIURL:      getEnv(envEvoAPIURL),
		EvoAPIKey:      getEnv(envEvoAPIKey),
		EvoAPIInstance: getEnv(envEvoAPIInstance),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	requiredFields := []envField{
		req(envDBHost, c.DB.Host),
		req(envDBPort, c.DB.Port),
		req(envDBUser, c.DB.User),
		req(envDBPassword, c.DB.Password),
		req(envDBName, c.DB.Name),
		req(envJWTSecret, c.JWTSecret),
		req(envEvoAPIURL, c.EvoAPIURL),
		req(envEvoAPIKey, c.EvoAPIKey),
		req(envEvoAPIInstance, c.EvoAPIInstance),
		req(envGroomPhone, c.Couple.Groom.Phone),
		req(envGroomURACF, c.Couple.Groom.URACF),
		req(envBridePhone, c.Couple.Bride.Phone),
		req(envBrideURACF, c.Couple.Bride.URACF),
	}

	var issues []string

	if missing := missingRequired(requiredFields); len(missing) > 0 {
		issues = append(issues, "missing required environment variables: "+strings.Join(missing, ", "))
	}

	if err := validatePort(envDBPort, c.DB.Port); err != nil {
		issues = append(issues, err.Error())
	}
	if err := validateDuration(envJWTExpiry, c.JWTExpiry); err != nil {
		issues = append(issues, err.Error())
	}
	if err := validateOneOf(envAppEnv, c.AppEnv, []string{"test", "production"}); err != nil {
		issues = append(issues, err.Error())
	}
	if err := validateEvoConfig(c.EvoAPIURL); err != nil {
		issues = append(issues, err.Error())
	}

	if len(issues) > 0 {
		return fmt.Errorf("invalid configuration:\n- %s", strings.Join(issues, "\n- "))
	}

	return nil
}

func req(name, value string) envField {
	return envField{name: name, value: value}
}

func missingRequired(fields []envField) []string {
	missing := make([]string, 0)
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			missing = append(missing, field.name)
		}
	}
	return missing
}

func validatePort(name, value string) error {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("%s must be a number between 1 and 65535", name)
	}
	return nil
}

func validateDuration(name, value string) error {
	if _, err := time.ParseDuration(value); err != nil {
		return fmt.Errorf("%s must be a valid duration (ex: 3h, 30m): %v", name, err)
	}
	return nil
}

func validateOneOf(name, value string, allowed []string) error {
	for _, item := range allowed {
		if value == item {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of: %s", name, strings.Join(allowed, ", "))
}

func validateEvoConfig(evoUrl string) error {
	if _, err := url.ParseRequestURI(evoUrl); err != nil {
		return fmt.Errorf("%s must be a valid URL: %v", envEvoAPIURL, err)
	}

	return nil
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func getEnvOrDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}
