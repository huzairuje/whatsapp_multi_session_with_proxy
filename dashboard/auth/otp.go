package auth

import (
	"crypto/rand"
	"fmt"
	"math/big" // For GenerateOTP

	"golang.org/x/crypto/bcrypt"
)

// GenerateOTP creates a random OTP string of digits of the given length.
func GenerateOTP(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("otp length must be positive")
	}

	otpChars := "0123456789"
	buffer := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(otpChars))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number for otp: %w", err)
		}
		buffer[i] = otpChars[num.Int64()]
	}
	return string(buffer), nil
}

// HashOTP generates a bcrypt hash of the OTP.
func HashOTP(otp string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash otp: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyOTP compares a hashed OTP with a plain OTP.
// Returns nil on success, or an error if they don't match (e.g., bcrypt.ErrMismatchedHashAndPassword).
func VerifyOTP(hashedOTP string, plainOTP string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedOTP), []byte(plainOTP))
	if err != nil {
		// This will return bcrypt.ErrMismatchedHashAndPassword if passwords don't match,
		// or other errors if something else went wrong during comparison.
		return fmt.Errorf("failed to verify otp: %w", err)
	}
	return nil
}
