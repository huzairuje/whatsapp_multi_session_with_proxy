package auth

import (
	"errors" // For errors.Is
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5" // Ensure this is the correct import for v5
)

const (
	testJwtSecretKey = "test-super-secret-key-for-jwt-12345abcde" // Made it a bit longer
	testAdminID      = uint(1)
	testPhoneNumber  = "1234567890"
)

func TestGenerateJWT(t *testing.T) {
	// Test successful token generation
	tokenString, err := GenerateJWT(testAdminID, testPhoneNumber, testJwtSecretKey, time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWT() with valid inputs error = %v", err)
	}
	if tokenString == "" {
		t.Error("GenerateJWT() with valid inputs returned empty token string")
	}

	// Test with empty secret key
	_, err = GenerateJWT(testAdminID, testPhoneNumber, "", time.Hour)
	if err == nil {
		t.Error("GenerateJWT() with empty secret key expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "jwt secret key cannot be empty") {
			t.Errorf("GenerateJWT() with empty secret key, expected error containing 'jwt secret key cannot be empty', got: %v", err)
		}
	}

	// Test with zero duration
	_, err = GenerateJWT(testAdminID, testPhoneNumber, testJwtSecretKey, 0)
	if err == nil {
		t.Error("GenerateJWT() with zero duration expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "jwt expiration time must be positive") {
			t.Errorf("GenerateJWT() with zero duration, expected error containing 'jwt expiration time must be positive', got: %v", err)
		}
	}

	// Test with negative duration (should also be invalid)
	_, err = GenerateJWT(testAdminID, testPhoneNumber, testJwtSecretKey, -time.Hour)
	if err == nil {
		t.Error("GenerateJWT() with negative duration expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "jwt expiration time must be positive") {
			t.Errorf("GenerateJWT() with negative duration, expected error containing 'jwt expiration time must be positive', got: %v", err)
		}
	}
}

func TestValidateJWT(t *testing.T) {
	// 1. Successful validation
	validToken, errGen := GenerateJWT(testAdminID, testPhoneNumber, testJwtSecretKey, time.Hour)
	if errGen != nil {
		t.Fatalf("Failed to generate token for TestValidateJWT: %v", errGen)
	}

	claims, err := ValidateJWT(validToken, testJwtSecretKey)
	if err != nil {
		t.Fatalf("ValidateJWT() with valid token error = %v", err)
	}
	if claims.AdminUserID != testAdminID {
		t.Errorf("ValidateJWT() claims.AdminUserID got = %v, want = %v", claims.AdminUserID, testAdminID)
	}
	if claims.PhoneNumber != testPhoneNumber {
		t.Errorf("ValidateJWT() claims.PhoneNumber got = %v, want = %v", claims.PhoneNumber, testPhoneNumber)
	}
	if claims.ExpiresAt == nil || claims.IssuedAt == nil {
		t.Errorf("ValidateJWT() token should have ExpiresAt and IssuedAt set")
	}
	if claims.Issuer != "whatsapp_multi_session_dashboard" {
		t.Errorf("ValidateJWT() claims.Issuer got = %v, want = 'whatsapp_multi_session_dashboard'", claims.Issuer)
	}

	// 2. Invalid secret key
	_, err = ValidateJWT(validToken, "wrong-secret-key-that-is-definitely-not-the-right-one")
	if err == nil {
		t.Error("ValidateJWT() with wrong secret key expected error, got nil")
	} else {
		// Check if the error is jwt.ErrSignatureInvalid or our wrapped error contains it.
		// The ValidateJWT function wraps errors, so we check for our wrapper and the underlying cause.
		if !strings.Contains(err.Error(), "failed to parse or validate jwt token") && !errors.Is(err, jwt.ErrSignatureInvalid) {
			t.Errorf("ValidateJWT() with wrong secret key, expected error containing 'failed to parse or validate jwt token' or specific signature error, got %v", err)
		}
	}


	// 3. Malformed token
	_, err = ValidateJWT("this.is.a.malformed.token", testJwtSecretKey)
	if err == nil {
		t.Error("ValidateJWT() with malformed token expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "failed to parse or validate jwt token") {
			t.Errorf("ValidateJWT() with malformed token, expected error containing 'failed to parse or validate jwt token', got %v", err)
		}
	}

	// 4. Expired token
	// Generate with a very short, positive duration for expiration testing
	expiredToken, errGenExp := GenerateJWT(testAdminID, testPhoneNumber, testJwtSecretKey, 1*time.Millisecond)
	if errGenExp != nil {
		t.Fatalf("Failed to generate expired token for TestValidateJWT: %v", errGenExp)
	}
	time.Sleep(50 * time.Millisecond) // Ensure token is definitely expired (increased sleep time)

	_, err = ValidateJWT(expiredToken, testJwtSecretKey)
	if err == nil {
		t.Error("ValidateJWT() with expired token expected error, got nil")
	} else {
		// Our ValidateJWT function wraps the error from jwt.ParseWithClaims.
		// The jwt library itself might return an error that wraps jwt.ErrTokenExpired.
		// So, we check if our wrapped error message contains the specific text of jwt.ErrTokenExpired.
		// Or, we can try to unwrap and check with errors.Is if our function preserves the chain.
		// Given the current ValidateJWT implementation: `fmt.Errorf("failed to parse or validate jwt token: %w", err)`
		// We can use errors.Is to check the wrapped error.
		if !errors.Is(err, jwt.ErrTokenExpired) {
			t.Errorf("ValidateJWT() with expired token, expected to wrap jwt.ErrTokenExpired, got %v. Underlying error: %v", err, errors.Unwrap(err))
		}
	}

	// 5. Empty token string
	_, err = ValidateJWT("", testJwtSecretKey)
	if err == nil {
		t.Error("ValidateJWT() with empty token string expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "jwt token string cannot be empty") {
			t.Errorf("ValidateJWT() with empty token string, expected error containing 'jwt token string cannot be empty', got: %v", err)
		}
	}

	// 6. Empty secret key for validation
	_, err = ValidateJWT(validToken, "")
	if err == nil {
		t.Error("ValidateJWT() with empty secret key expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "jwt secret key cannot be empty") {
			t.Errorf("ValidateJWT() with empty secret key, expected error containing 'jwt secret key cannot be empty', got: %v", err)
		}
	}
}
