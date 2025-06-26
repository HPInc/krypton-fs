package common

import (
	"crypto/rand"
	"encoding/base32"
	"math/big"
	"strings"
)

const (
	defaultContent = "stock default content"
)

// get random bytes of english lowercase letters
// up to a max length
// note that errors are silently suppressed for default behavior
// not suitable for production.
func NewRandomString(maxLength int) string {
	n64Length := int64(maxLength)
	length, err := rand.Int(rand.Reader, big.NewInt(n64Length))
	if err != nil {
		length = big.NewInt(n64Length)
	}
	strLength := length.Uint64()
	bytes := make([]byte, strLength)
	if _, err = rand.Read(bytes); err != nil {
		return defaultContent
	}
	return strings.ToLower(base32.StdEncoding.EncodeToString(bytes)[:strLength])
}
