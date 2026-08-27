package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"megin/internal/config"
	"net/url"
	"strings"
	"time"
)

const (
	totpSecretBytes = 20
	totpCodeDigits  = 6
	totpStepSeconds = 30
	totpAllowedSkew = 1
	totpNonceSize   = 12
)

func generateTOTPSecret() (string, error) {
	buf := make([]byte, totpSecretBytes)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(buf), "="), nil
}

func encryptTOTPSecret(secret string) (string, error) {
	key := totpCipherKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, totpNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(secret), nil)
	payload := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(payload), nil
}

func decryptTOTPSecret(ciphertext string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	key := totpCipherKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("invalid totp secret payload")
	}
	nonce, payload := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func totpCipherKey() []byte {
	sum := sha256.Sum256([]byte(config.GetConfig().Jwt.Secret))
	return sum[:]
}

func totpCodeAt(secret string, now time.Time) (string, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	decoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	key, err := decoder.DecodeString(secret)
	if err != nil {
		return "", err
	}

	counter := uint64(now.Unix() / totpStepSeconds)
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], counter)

	mac := hmac.New(sha1.New, key)
	if _, err := mac.Write(msg[:]); err != nil {
		return "", err
	}
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(sum[offset : offset+4])
	truncated &= 0x7fffffff
	code := truncated % uint32(math.Pow10(totpCodeDigits))
	return fmt.Sprintf("%0*d", totpCodeDigits, code), nil
}

func verifyTOTPCode(secret string, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpCodeDigits {
		return false
	}
	for step := -totpAllowedSkew; step <= totpAllowedSkew; step++ {
		candidate, err := totpCodeAt(secret, now.Add(time.Duration(step*totpStepSeconds)*time.Second))
		if err == nil && candidate == code {
			return true
		}
	}
	return false
}

func buildTOTPAuthURL(issuer, account, secret string) string {
	label := url.PathEscape(issuer + ":" + account)
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("issuer", issuer)
	values.Set("algorithm", "SHA1")
	values.Set("digits", fmt.Sprintf("%d", totpCodeDigits))
	values.Set("period", fmt.Sprintf("%d", totpStepSeconds))
	return "otpauth://totp/" + label + "?" + values.Encode()
}
