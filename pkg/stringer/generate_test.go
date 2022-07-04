//go:build unit

package stringer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCode(t *testing.T) {
	numberOfCodes := 10000
	codes := make(map[string]string, numberOfCodes)
	var collisions int
	for i := 0; i < numberOfCodes; i++ {
		result := GenerateCode([]rune("0123456789ABCDEFGHIJKLMNOPQRSTUVXWYZ"), 6)
		if _, ok := codes[result]; ok {
			collisions++
		}
		codes[result] = result
	}

	assert.Conditionf(
		t,
		func() bool {
			return collisions < int(float64(numberOfCodes)*0.01)
		},
		"collision greater than 1%d at",
		collisions,
	)

	t.Logf(
		"codes generated: %d, percentage of collisions: %.2f%%",
		len(codes),
		(float64(collisions)*100)/float64(numberOfCodes),
	)
}

func BenchmarkGenerateCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateCode([]rune("0123456789ABCDEFGHIJKLMNOPQRSTUVXWYZ"), 6)
	}
}
