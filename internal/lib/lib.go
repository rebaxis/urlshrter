package lib

import (
	"crypto/rand"
	"math/big"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var alphabetLen = big.NewInt(int64(len(alphabet)))

func GenerateRandomAlphabetString(length int) string {
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			panic("failed to generate random number: " + err.Error())
		}
		b[i] = alphabet[num.Int64()]
	}
	return string(b)
}
