package toolkit

import (
	"crypto/rand"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type used to instantiate the module. Any variable of this type will have access
// to the methods with the receiver *Tools
type Tools struct{}

// RandomString returns a random character of the specified length
func (t *Tools) RandomString(length int) string {
	// make a slice of runes of the specified length
	s := make([]rune, length)
	// convert the randomStringSource string to a slice of runes
	r := []rune(randomStringSource)

	for i := range s {
		// rand.Prime is used to generate a random prime number of the specified length
		p, _ := rand.Prime(rand.Reader, len(r))

		// convert the prime number to a uint64 and use it as the index of the randomStringSource
		x, y := p.Uint64(), uint64(len(r))

		s[i] = r[x%y]
	}

	return string(s)
}
