package config

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	t.Run("Should not return error for valid config", testValidateValidConfig)
	t.Run("Should return error for missing required fields", testValidateMissingRequiredFields)
	t.Run("Should return error for invalid formats", testValidateInvalidFormats)
	t.Run("Should return error for invalid EVO URL", testValidateEvoInvalidURL)
}

func testValidateValidConfig(t *testing.T) {
	cfg := validConfig()

	if err := cfg.validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func testValidateMissingRequiredFields(t *testing.T) {
	cfg := validConfig()
	cfg.DB.Host = ""
	cfg.DB.User = ""
	cfg.JWTSecret = ""
	cfg.Couple.Groom.Phone = ""
	cfg.Couple.Bride.URACF = ""

	err := cfg.validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	wantFields := []string{
		envDBHost,
		envDBUser,
		envJWTSecret,
		envGroomPhone,
		envBrideURACF,
	}
	for _, field := range wantFields {
		if !strings.Contains(err.Error(), field) {
			t.Errorf("expected error to contain %q, got: %v", field, err)
		}
	}
}

func testValidateInvalidFormats(t *testing.T) {
	cfg := validConfig()
	cfg.DB.Port = "abc"
	cfg.JWTExpiry = "never"
	cfg.AppEnv = "staging"

	err := cfg.validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	wantSnippets := []string{
		envDBPort + " must be a number between 1 and 65535",
		envJWTExpiry + " must be a valid duration",
		envAppEnv + " must be one of",
	}
	for _, snippet := range wantSnippets {
		if !strings.Contains(err.Error(), snippet) {
			t.Errorf("expected error to contain %q, got: %v", snippet, err)
		}
	}
}

func testValidateEvoInvalidURL(t *testing.T) {
	cfg := validConfig()
	cfg.EvoAPIURL = "not-a-url"
	cfg.EvoAPIKey = "secret"
	cfg.EvoAPIInstance = "instance"

	err := cfg.validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), envEvoAPIURL+" must be a valid URL") {
		t.Fatalf("expected EVO URL validation error, got: %v", err)
	}
}

func validConfig() Config {
	return Config{
		DB: DBConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			Name:     "parasempre",
			SSLMode:  "disable",
		},
		CORSOrigin: "http://localhost:3000",
		AppEnv:     "test",
		Couple: CoupleConfig{
			Groom: CoupleUserConfig{
				Phone: "+5511999999999",
				URACF: "groom-uracf",
			},
			Bride: CoupleUserConfig{
				Phone: "+5511888888888",
				URACF: "bride-uracf",
			},
		},
		JWTSecret:      "secret",
		JWTExpiry:      "3h",
		EvoAPIURL:      "http://localhost:8081",
		EvoAPIKey:      "secret",
		EvoAPIInstance: "instance",
	}
}
