package random

import (
	"math/rand"
	"sync"
	"time"
)

var (
	rnd  *rand.Rand
	once sync.Once
)

func initRandom() {
	once.Do(func() {
		rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	})
}

func NewRandomString(size int) string {
	initRandom()

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	b := make([]rune, size)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}

	return string(b)
}
