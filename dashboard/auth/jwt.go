package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Note: In a real application, the secret key for JWT generation and validation
// should come from a secure configuration source, such as environment variables,
// and should be a strong, randomly generated string.
// For the purpose of this implementation, the secret key will be passed as an argument
// to the GenerateJWT and ValidateJWT functions. This avoids hardcoding it globally
// or relying on a configuration mechanism that isn't part of this specific subtask.

// JWTClaims defines the structure of the JWT claims, including custom ones.
type JWTClaims struct {
	AdminUserID uint   `json:"admin_user_id"`
	PhoneNumber string `json:"phone_number"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token for an admin user.
// secretKey: The secret key used to sign the token.
// expirationTime: Duration for which the token will be valid.
func GenerateJWT(adminUserID uint, phoneNumber string, secretKey string, expirationTime time.Duration) (string, error) {
	if secretKey == "" {
		return "", fmt.Errorf("jwt secret key cannot be empty")
	}
	if expirationTime <= 0 {
		return "", fmt.Errorf("jwt expiration time must be positive")
	}

	expiration := time.Now().Add(expirationTime)
	claims := &JWTClaims{
		AdminUserID: adminUserID,
		PhoneNumber: phoneNumber,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "whatsapp_multi_session_dashboard", // Optional: an issuer name
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign jwt token: %w", err)
	}
	return signedToken, nil
}

// ValidateJWT parses and validates a JWT token string.
// Returns the claims if the token is valid, otherwise returns an error.
// secretKey: The secret key used to validate the token's signature.
func ValidateJWT(tokenString string, secretKey string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("jwt token string cannot be empty")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("jwt secret key cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		// This error could be due to various reasons like expired token, malformed token, signature mismatch, etc.
		// jwt.ErrTokenExpired, jwt.ErrTokenNotValidYet, jwt.ErrSignatureInvalid are some specific errors from the library.
		return nil, fmt.Errorf("failed to parse or validate jwt token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	// This case should ideally not be reached if err from ParseWithClaims is handled correctly,
	// as token.Valid would be false and ParseWithClaims would return an error.
	// However, it's a fallback.
	return nil, fmt.Errorf("invalid jwt token or claims type assertion failed")
}
