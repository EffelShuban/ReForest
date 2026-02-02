package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateAndValidateToken_Success(t *testing.T) {
	provider := NewJWTProvider("secret-key")
	userID := uuid.New()
	role := "ADMIN"

	token, err := provider.GenerateToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := provider.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Fatalf("expected role %s, got %s", role, claims.Role)
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	userID := uuid.New()

	// Token signed with a different secret
	wrongProvider := NewJWTProvider("wrong-secret")
	token, err := wrongProvider.GenerateToken(userID, "ADMIN")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	provider := NewJWTProvider("correct-secret")
	if _, err := provider.ValidateToken(token); err == nil {
		t.Fatalf("expected signature validation error, got nil")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	provider := NewJWTProvider("secret-key")
	userID := uuid.New()

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"role":    "ADMIN",
		"exp":     time.Now().Add(-1 * time.Hour).Unix(),
	})

	tokenString, err := expiredToken.SignedString([]byte("secret-key"))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	if _, err := provider.ValidateToken(tokenString); err == nil {
		t.Fatalf("expected expired token error, got nil")
	}
}

func TestValidateToken_MissingClaims(t *testing.T) {
	provider := NewJWTProvider("secret-key")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": uuid.New().String(),
		// "role" omitted intentionally
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte("secret-key"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if _, err := provider.ValidateToken(tokenString); err == nil {
		t.Fatalf("expected missing role claim error, got nil")
	}
}
