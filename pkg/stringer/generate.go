package stringer

import (
	"crypto/rand"
	"math/big"
	"strings"
)

func GenerateCode(variants []rune, size int) string {
	var sb strings.Builder
	newInt := big.NewInt(int64(len(variants)))
	for sb.Len() < size {
		n, err := rand.Int(rand.Reader, newInt)
		if err == nil {
			sb.WriteRune(variants[n.Int64()])
		}
	}

	return sb.String()
}
