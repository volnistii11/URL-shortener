package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var urlMap map[string]string

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}
		r.Body.Close()

		urlMap[string(body)] = r.Host + r.RequestURI + randString(10)

		w.Write([]byte(urlMap[string(body)]))
	}
}

func main() {
	urlMap = map[string]string{}

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainPage)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
