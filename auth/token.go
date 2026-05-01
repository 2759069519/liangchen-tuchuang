package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

var tokenSecret []byte

func InitTokenSecret() {
	b := make([]byte, 32)
	rand.Read(b)
	tokenSecret = b
}

// GenerateToken creates a signed token with username embedded.
// Format: hex(timestamp).username.expiry.signature
func GenerateToken(username string) (string, error) {
	expiry := time.Now().Add(24 * time.Hour).Unix()
	payload := hex.EncodeToString([]byte(time.Now().Format("2006-01-02T15:04"))) + "." + username + "." + intToStr(expiry)
	mac := hmac.New(sha256.New, tokenSecret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

// ValidateToken checks signature and expiry, returns (valid, username).
// For backward compatibility with old 3-part tokens, username returns "admin".
func ValidateToken(tokenStr string) (bool, string) {
	parts := strings.SplitN(tokenStr, ".", 4)
	if len(parts) == 3 {
		// Old format: hex(time).expiry.sig (no username)
		payload := parts[0] + "." + parts[1]
		sig := parts[2]
		mac := hmac.New(sha256.New, tokenSecret)
		mac.Write([]byte(payload))
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(sig), []byte(expected)) {
			return false, ""
		}
		expiry := strToInt(parts[1])
		if time.Now().Unix() > expiry {
			return false, ""
		}
		return true, "admin"
	}
	if len(parts) == 4 {
		// New format: hex(time).username.expiry.sig
		payload := parts[0] + "." + parts[1] + "." + parts[2]
		sig := parts[3]
		mac := hmac.New(sha256.New, tokenSecret)
		mac.Write([]byte(payload))
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(sig), []byte(expected)) {
			return false, ""
		}
		expiry := strToInt(parts[2])
		if time.Now().Unix() > expiry {
			return false, ""
		}
		return true, parts[1]
	}
	return false, ""
}

func intToStr(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func strToInt(s string) int64 {
	var n int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		}
	}
	return n
}
