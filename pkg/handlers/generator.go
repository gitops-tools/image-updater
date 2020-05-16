package hook

import (
	"fmt"
	"math/rand"
	"time"
)

var timeSeed = rand.New(rand.NewSource(time.Now().UnixNano()))

type randomNameGenerator struct {
	rand *rand.Rand
}

// Generates a name from the prefix with an additional 5 random alphabetic
// characters.
// TODO: this should limit the length based on the prefix because branch names
// have a limit.
func (r randomNameGenerator) prefixedName(prefix string) string {
	charset := "abcdefghijklmnopqrstuvwyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[r.rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s%s", prefix, b)
}
