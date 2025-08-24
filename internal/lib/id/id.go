package id

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a 32-hex-character unique identifier (128-bit random).
func New() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
