package auth

import (
	"testing"
	"time"
)

func TestJWTGenerateAndParse(t *testing.T) {
	svc := NewJWTService("test-secret-key-for-testing", 1*time.Hour)

	token, err := svc.Generate(42, "USR01", "guest")
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := svc.Parse(token)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("expected user_id 42, got %d", claims.UserID)
	}
	if claims.URACF != "USR01" {
		t.Fatalf("expected uracf USR01, got %q", claims.URACF)
	}
	if claims.Role != "guest" {
		t.Fatalf("expected role guest, got %q", claims.Role)
	}
}

func TestJWTExpired(t *testing.T) {
	svc := NewJWTService("test-secret", -1*time.Hour)

	token, err := svc.Generate(1, "USR01", "guest")
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}

	_, err = svc.Parse(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWTInvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour)

	_, err := svc.Parse("invalid.token.string")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestJWTWrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-1", 1*time.Hour)
	svc2 := NewJWTService("secret-2", 1*time.Hour)

	token, _ := svc1.Generate(1, "USR01", "guest")
	_, err := svc2.Parse(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}
