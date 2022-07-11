package random

import (
	"math/rand"
	"time"
)

const charset = "aBcD123.!#"

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

// StringWithCharset generates random string of provided length
// using provided charset
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[randSource.Intn(len(charset))]
	}
	return string(b)
}

// String generates random string of provided length
// using default charset
//
// To generate random string using different charset
// use StringWithCharset function
func String(length int) string {
	return StringWithCharset(length, charset)
}
