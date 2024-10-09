package hashutils

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256 calculates schema hash using SHA256 algorithm.
func SHA256(data []byte) string {
	hasher := sha256.New()
	_, _ = hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}
