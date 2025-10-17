package lib

import (
	"math/rand"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRandomAlphabetString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[seededRand.Intn(len(alphabet))]
	}
	return string(b)
}

// Функция для проверки существования значения в map
func checkForValue(url string, urlsMap map[string]string) bool {
	//traverse through the map
	for _, value := range urlsMap {
		//check if present value is equals to userValue
		if value == url {
			//if same return true
			return true
		}
	}
	return false
}
