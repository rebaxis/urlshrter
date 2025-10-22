package lib

import (
	"math/rand"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateRandomAlphabetString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[seededRand.Intn(len(alphabet))]
	}
	return string(b)
}

// Функция для проверки существования значения в map
func CheckForValue(url string, urlsMap map[string]string) bool {
	idMap := make(map[string]string)
	for key, value := range urlsMap {
		idMap[value] = key
	}

	_, exists := idMap[url]
	if exists {
		return true
	} else {
		return false
	}
}
