package rewriter

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const cchSalt = "59cf53e54c78"

var (
	cchPositions = []int{4, 7, 20}
	fallbackHash string
)

func init() {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	fallbackHash = hex.EncodeToString(b)[:3]
}

func computeCCH(firstUserMsg, version string) string {
	var chars [3]byte
	for i, pos := range cchPositions {
		if pos < len(firstUserMsg) {
			chars[i] = firstUserMsg[pos]
		} else {
			chars[i] = '0'
		}
	}
	hash := sha256.Sum256([]byte(cchSalt + string(chars[:]) + version))
	return hex.EncodeToString(hash[:])[:3]
}
