package authentication

import (
	"strings"
	"testing"
)

func TestHashPasswordCreatesVerifiableArgon2idPHCString(t *testing.T) {
	encoded, err := hashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("unexpected password format: %q", encoded)
	}
	matched, legacy := verifyPassword("correct horse battery staple", encoded)
	if !matched || legacy {
		t.Fatalf("Argon2id password verification = (%v, %v), want (true, false)", matched, legacy)
	}
	matched, legacy = verifyPassword("wrong", encoded)
	if matched || legacy {
		t.Fatalf("wrong password verification = (%v, %v), want (false, false)", matched, legacy)
	}
}

func TestVerifyPasswordRecognizesLegacyVerifier(t *testing.T) {
	stored := SHA256("legacy-password", "ignored-salt")
	matched, legacy := verifyPassword("legacy-password", stored)
	if !matched || !legacy {
		t.Fatalf("legacy verification = (%v, %v), want (true, true)", matched, legacy)
	}
}

func TestVerifyPasswordRejectsMalformedOrExcessiveArgon2id(t *testing.T) {
	values := []string{
		"$argon2id$broken",
		"$argon2id$v=19$m=1048576,t=2,p=1$YWJjZGVmZ2hpamtsbW5vcA$YWJj",
		"$argon2id$v=18$m=19456,t=2,p=1$YWJjZGVmZ2hpamtsbW5vcA$YWJj",
	}
	for _, stored := range values {
		matched, legacy := verifyPassword("password", stored)
		if matched || legacy {
			t.Fatalf("malformed verifier %q returned (%v, %v)", stored, matched, legacy)
		}
	}
}
