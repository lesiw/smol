package randstr

import (
	"crypto/rand"
	"math/big"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func New(length int) (string, error) {
	id := make([]byte, length)
	for i := range id {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		id[i] = alphabet[num.Int64()]
	}
	return string(id), nil
}
