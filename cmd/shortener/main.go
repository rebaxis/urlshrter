package main

import (
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"slices"
	"time"
)

var urlS = make(map[string]string)

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

func screateId(res http.ResponseWriter, req *http.Request) {
	// Проверяем что метод POST
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	// Проверяем что Content-Type содержит text/plain
	contTypeHeader := req.Header.Values("Content-Type")
	if !slices.Contains(contTypeHeader, "text/plain") {
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
	res.Write([]byte(req.URL.Scheme + req.Host + req.RequestURI + shortSt))
}

func getUrlById(res http.ResponseWriter, req *http.Request) {
	// Проверяем что Content-Type содержит text/plain
	contTypeHeader := req.Header.Values("Content-Type")
	if !slices.Contains(contTypeHeader, "text/plain") {
		http.Error(res, "Content-Type header must be text/plain", http.StatusBadRequest)
		return
	}

	// Возвращаем URL
	idString := req.PathValue("id")
	if pURL, ok := urlS[idString]; ok {
		res.Header().Add("Location", pURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
		res.Write([]byte(""))
	} else {
		http.Error(res, "There isn't URL with this id", http.StatusBadRequest)
		return
	}

}

func main() {
	// добавить middleware text/plain

	mux := http.NewServeMux()
	mux.HandleFunc("/", screateId)
	mux.HandleFunc("GET /{id}", getUrlById)

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		panic(err)
	}
}
