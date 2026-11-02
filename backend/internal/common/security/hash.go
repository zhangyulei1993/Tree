package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
)

var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

func ValidPhone(phone string) bool {
	return phonePattern.MatchString(phone)
}

func HashPlain(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func PhoneHash(phone string) string {
	return HashPlain(phone)
}

func GenerateNumericCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}
	max := big.NewInt(1)
	for i := 0; i < length; i++ {
		max.Mul(max, big.NewInt(10))
	}
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", length, value.Int64()), nil
}

func HashVerificationCode(phone string, scene string, code string) (string, error) {
	return HashPassword(phone + ":" + scene + ":" + code)
}

func CheckVerificationCode(phone string, scene string, code string, codeHash string) bool {
	return CheckPassword(phone+":"+scene+":"+code, codeHash)
}
