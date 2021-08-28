package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	myParam := r.URL.Query().Get("param")
	if myParam != "" {
		fmt.Println("myParam is", myParam)
		//todo: здесь мы должны возвращать ошибку, если здесь будет пусто
	}
	key := r.FormValue("key")
	if key != "" {
		fmt.Println("key is", key)
	}
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("starting server at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
