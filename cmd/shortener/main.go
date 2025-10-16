package main

import (
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// define storage variable with element for testing
var urlS = map[string]string{"testTest": "http://test-test.test"}

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

func screateID(res http.ResponseWriter, req *http.Request) {
	// Проверяем что метод POST
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	// Проверяем что Content-Type содержит text/plain
	contTypeHeader := req.Header.Values("Content-Type")
	contTypeCorrect := false
	for _, contHead := range contTypeHeader {
		if strings.Contains(contHead, "text/plain") {
			contTypeCorrect = true
		}
	}
	if !contTypeCorrect {
		http.Error(res, "Content-Type header must be text/plain", http.StatusBadRequest)
		return
	}

	// Проверяем что тело запроса не пустое
	body, err := io.ReadAll(req.Body)
	if err != nil || string(body) == "" {
		http.Error(res, "Body is empty!", http.StatusBadRequest)
		return
	}

	sBody := string(body)

	// Проверяем URL на валидность
	pURL, err := url.ParseRequestURI(sBody)
	if err != nil {
		http.Error(res, "Body has invalid URL", http.StatusBadRequest)
		return
	}

	// Проверяем что этот URL еще не добавлен
	if checkForValue(pURL.String(), urlS) {
		http.Error(res, "This URL already has short name", http.StatusBadRequest)
		return
	}

	// Добавляем URL в наш map
	shortSt := generateRandomAlphabetString(8)
	urlS[shortSt] = pURL.String()

	// Возвращаем id ссылки
	res.Header().Add("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://" + req.Host + req.RequestURI + shortSt))
}

func getURLByID(res http.ResponseWriter, req *http.Request) {
	// Возвращаем URL
	idString := req.PathValue("id")
	if pURL, ok := urlS[idString]; ok {
		res.Header().Add("Location", pURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
		return
	} else {
		http.Error(res, "There isn't URL with this id "+idString, http.StatusBadRequest)
		return
	}

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", screateID)
	mux.HandleFunc("GET /{id}", getURLByID)

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		panic(err)
	}
}
