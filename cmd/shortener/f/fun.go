package fun

import (
	"math/rand"
	"shortingURL/cmd/shortener/config"
	"strings"
)

var SequenceUUID uint

// Генератор хеша. Использует константу hashLen для определения длины
func GetHash() string {
	sb := strings.Builder{}
	sb.Grow(config.HashLen)
	for i := 0; i < config.HashLen; i++ {
		sb.WriteByte(config.Charset[rand.Intn(len(config.Charset))])
	}
	return sb.String()
}

// Генератор следующего uuid для базы урлов.
func NextSequenceID() uint {
	SequenceUUID += 1
	return SequenceUUID
}
