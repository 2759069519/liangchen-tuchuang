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

func GenerateToken() (string, error) {
	expiry := time.Now().Add(24 * time.Hour).Unix()
	payload := hex.EncodeToString([]byte(time.Now().Format("2006-01-02T15:04"))) + "." + intToStr(expiry)
	mac := hmac.New(sha256.New, tokenSecret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

func ValidateToken(tokenStr string) bool {
	parts := strings.SplitN(tokenStr, ".", 3)
	if len(parts) != 3 { return false }
	payload := parts[0] + "." + parts[1]
	sig := parts[2]
	mac := hmac.New(sha256.New, tokenSecret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) { return false }
	expiry := strToInt(parts[1])
	return time.Now().Unix() <= expiry
}

func intToStr(n int64) string {
	if n == 0 { return "0" }
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
