package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HashAssistanceCode is deterministic only inside this deployment. An 8-digit
// code must be looked up by hash, therefore a plain SHA-256 hash is not enough.
func HashAssistanceCode(secret, code string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}
