package handlers

import (
	"fmt"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	storage.URLMap = map[string]string{}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		CreateShortURL(w, r)
	} else if r.Method == http.MethodGet {
		GetFullURL(w, r)
	}
}

func CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()
	if len(body) == 0 {
		http.Error(w, "Body is empty.", http.StatusBadRequest)
		return
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	shortURL := randString(10)
	storage.URLMap[shortURL] = string(body)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%v://%v%v%v", scheme, r.Host, r.RequestURI, shortURL)))
}

func GetFullURL(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	shortURL := strings.Split(path, "/")[0]

	if fullURL, ok := storage.URLMap[shortURL]; ok {
		w.Header().Set("Location", fullURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
