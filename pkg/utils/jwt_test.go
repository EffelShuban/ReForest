package utils

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestJWTProvider_GenerateAndValidateToken(t *testing.T) {
	secret := "test-secret"
	provider := NewJWTProvider(secret)

	userID := uuid.New()
	role := "ADMIN"

	token, err := provider.GenerateToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := provider.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("UserID mismatch: got %s want %s", claims.UserID, userID)
	}
	if claims.Role != role {
		t.Fatalf("Role mismatch: got %s want %s", claims.Role, role)
	}
}

func TestJWTProvider_ValidateToken_InvalidSignature(t *testing.T) {
	secret := "test-secret"
	provider := NewJWTProvider(secret)

	userID := uuid.New().String()
	role := "SPONSOR"

	// Token signed with a different secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
	})
	badToken, err := token.SignedString([]byte("other-secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := provider.ValidateToken(badToken); err == nil {
		t.Fatal("ValidateToken() expected error for invalid signature, got nil")
	}
}

func TestJWTProvider_ValidateToken_MissingUserID(t *testing.T) {
	provider := NewJWTProvider("secret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "ADMIN",
	})
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := provider.ValidateToken(signed); err == nil {
		t.Fatal("expected error for missing user_id claim, got nil")
	}
}

func TestJWTProvider_ValidateToken_InvalidUUID(t *testing.T) {
	provider := NewJWTProvider("secret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "not-a-uuid",
		"role":    "ADMIN",
	})
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := provider.ValidateToken(signed); err == nil {
		t.Fatal("expected error for invalid user_id format, got nil")
	}
}

func TestJWTProvider_ValidateToken_MissingRole(t *testing.T) {
	provider := NewJWTProvider("secret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": uuid.New().String(),
	})
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := provider.ValidateToken(signed); err == nil {
		t.Fatal("expected error for missing role claim, got nil")
	}
}
