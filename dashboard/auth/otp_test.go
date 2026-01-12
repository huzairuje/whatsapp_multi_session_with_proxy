package auth

import (
	"errors" // For errors.Is
	"regexp"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGenerateOTP(t *testing.T) {
	// Test valid length
	validLength := 6
	otp, err := GenerateOTP(validLength)
	if err != nil {
		t.Fatalf("GenerateOTP(%d) returned error: %v", validLength, err)
	}
	if len(otp) != validLength {
		t.Errorf("Expected OTP length %d, got %d for OTP: %s", validLength, len(otp), otp)
	}
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(otp) {
		t.Errorf("Expected OTP to contain only digits, got %s", otp)
	}

	// Test another valid length
	anotherValidLength := 8
	otp, err = GenerateOTP(anotherValidLength)
	if err != nil {
		t.Fatalf("GenerateOTP(%d) returned error: %v", anotherValidLength, err)
	}
	if len(otp) != anotherValidLength {
		t.Errorf("Expected OTP length %d, got %d for OTP: %s", anotherValidLength, len(otp), otp)
	}
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(otp) {
		t.Errorf("Expected OTP to contain only digits, got %s", otp)
	}

	// Test zero length
	_, err = GenerateOTP(0)
	if err == nil {
		t.Error("GenerateOTP(0) expected error, got nil")
	} else {
		// Optional: Check for a specific error message if your GenerateOTP returns a specific one
		if !strings.Contains(err.Error(), "otp length must be positive") {
			t.Errorf("GenerateOTP(0) expected error containing 'otp length must be positive', got: %v", err)
		}
	}


	// Test negative length
	_, err = GenerateOTP(-1)
	if err == nil {
		t.Error("GenerateOTP(-1) expected error, got nil")
	} else {
		if !strings.Contains(err.Error(), "otp length must be positive") {
			t.Errorf("GenerateOTP(-1) expected error containing 'otp length must be positive', got: %v", err)
		}
	}
}

func TestHashAndVerifyOTP(t *testing.T) {
	// Use a dynamically generated OTP for more robust testing
	plainOTP, err := GenerateOTP(6)
	if err != nil {
		t.Fatalf("Failed to generate OTP for testing HashAndVerifyOTP: %v", err)
	}

	hashedOTP, err := HashOTP(plainOTP)
	if err != nil {
		t.Fatalf("HashOTP(%s) returned error: %v", plainOTP, err)
	}
	if hashedOTP == "" {
		t.Fatalf("HashOTP(%s) returned empty hash", plainOTP)
	}

	// Test successful verification
	err = VerifyOTP(hashedOTP, plainOTP)
	if err != nil {
		t.Errorf("VerifyOTP() with correct OTP ('%s') returned error: %v", plainOTP, err)
	}

	// Test incorrect OTP
	incorrectOTP := plainOTP + "0" // Make it incorrect
	if plainOTP == "999999" { // Edge case if generated OTP is all 9s
		incorrectOTP = "000000"
	}

	err = VerifyOTP(hashedOTP, incorrectOTP)
	if err == nil {
		t.Errorf("VerifyOTP() with incorrect OTP ('%s' vs original '%s') expected error, got nil", incorrectOTP, plainOTP)
	}
	// Use errors.Is for checking specific error types from libraries
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Errorf("VerifyOTP() with incorrect OTP expected bcrypt.ErrMismatchedHashAndPassword, got %T: %v", err, err)
	}

	// Test verification with an empty plain OTP
	err = VerifyOTP(hashedOTP, "")
	if err == nil {
		t.Error("VerifyOTP() with empty plain OTP expected error, got nil")
	}
	// bcrypt.CompareHashAndPassword against an empty string and a valid hash should result in ErrMismatchedHashAndPassword
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Errorf("VerifyOTP() with empty plain OTP expected bcrypt.ErrMismatchedHashAndPassword, got %T: %v", err, err)
	}
    
    // Test verification with an empty hashed OTP
    err = VerifyOTP("", plainOTP)
	if err == nil {
		t.Error("VerifyOTP() with empty hashed OTP expected error, got nil")
	}
	// The error here will likely be "hashedSecret too short to be a bcrypted password" or similar from bcrypt
	// or our wrapper error "failed to verify otp: crypto/bcrypt: hashedSecret too short to be a bcrypted password"
	if !strings.Contains(err.Error(), "hashedSecret too short") && !strings.Contains(err.Error(), "failed to verify otp") {
		t.Errorf("VerifyOTP() with empty hashed OTP expected specific error related to short hash, got %v", err)
	}


    // Test hashing an empty string (bcrypt allows this, so HashOTP should too)
    emptyPlainOTP := ""
    emptyHashedOTP, err := HashOTP(emptyPlainOTP)
    if err != nil {
        t.Fatalf("HashOTP(\"\") returned error: %v", err)
    }
    if emptyHashedOTP == "" {
		t.Fatal("HashOTP(\"\") returned empty hash")
	}

    // Test successful verification of an empty string with its hash
    err = VerifyOTP(emptyHashedOTP, emptyPlainOTP)
    if err != nil {
        t.Errorf("VerifyOTP() with empty string and its hash returned error: %v", err)
    }

	// Test verification of non-empty OTP against hash of empty string
	err = VerifyOTP(emptyHashedOTP, "notempty")
	if err == nil {
        t.Error("VerifyOTP() with non-empty OTP against hash of empty string expected error, got nil")
    }
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Errorf("VerifyOTP() with non-empty OTP vs empty hash expected bcrypt.ErrMismatchedHashAndPassword, got %T: %v", err, err)
	}
}
