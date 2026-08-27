package service

import (
	"megin/internal/config"
	"strings"
	"testing"
	"time"
)

func TestTOTPEncryptDecryptRoundTrip(t *testing.T) {
	config.InitConfig("../../../config/config-test.yaml", config.RunModeMixed)

	secret, err := generateTOTPSecret()
	if err != nil {
		t.Fatalf("generateTOTPSecret error = %v", err)
	}
	encrypted, err := encryptTOTPSecret(secret)
	if err != nil {
		t.Fatalf("encryptTOTPSecret error = %v", err)
	}
	decrypted, err := decryptTOTPSecret(encrypted)
	if err != nil {
		t.Fatalf("decryptTOTPSecret error = %v", err)
	}
	if decrypted != secret {
		t.Fatalf("decrypted secret = %q, want %q", decrypted, secret)
	}
}

func TestVerifyTOTPCode(t *testing.T) {
	config.InitConfig("../../../config/config-test.yaml", config.RunModeMixed)

	now := time.Unix(1_700_000_000, 0)
	secret := "JBSWY3DPEHPK3PXP"
	code, err := totpCodeAt(secret, now)
	if err != nil {
		t.Fatalf("totpCodeAt error = %v", err)
	}
	if !verifyTOTPCode(secret, code, now) {
		t.Fatalf("verifyTOTPCode should accept current code")
	}
	if !verifyTOTPCode(secret, code, now.Add(30*time.Second)) {
		t.Fatalf("verifyTOTPCode should accept code within skew window")
	}
	if verifyTOTPCode(secret, "000000", now) && code != "000000" {
		t.Fatalf("verifyTOTPCode accepted invalid code")
	}
}

func TestBuildTOTPAuthURL(t *testing.T) {
	got := buildTOTPAuthURL("xadmin", "admin", "ABC123")
	if !strings.HasPrefix(got, "otpauth://totp/") {
		t.Fatalf("auth url prefix = %q", got)
	}
	if !strings.Contains(got, "secret=ABC123") {
		t.Fatalf("auth url missing secret: %q", got)
	}
	if !strings.Contains(got, "issuer=xadmin") {
		t.Fatalf("auth url missing issuer: %q", got)
	}
}
