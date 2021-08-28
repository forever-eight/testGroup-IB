package main

import (
	"container/list"
	"fmt"
	"log"
	"net/http"
)

var cont map[string]*list.List

// Добавление в очередь
func addToQueue(name string, param string) {
	if cont[name] == nil {
		queue := list.New()
		queue.PushBack(param)
		cont[name] = queue
	} else {
		cont[name].PushBack(param)
	}

}

// Изъятие из очереди
func getFromQueue(name string) string {
	answer, ok := cont[name]
	if !ok {
		log.Println("Problem with map")
		return ""
	}

	if answer.Len() > 0 {
		e := answer.Front() // Первый элемент
		defer answer.Remove(e)

		val := e.Value.(string)
		log.Println("val:", val)
		return val
	}

	return ""
}

// Узнать, каким методом нам посылается запрос
func Choice(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		get(w, r)
	} else if r.Method == http.MethodPut {
		put(w, r)
	}
}

// Получаем наше значение и добавляем в очередь
func put(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	v := r.URL.Query().Get("v")
	addToQueue(name, v)
	if v == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	fmt.Println(name)
	fmt.Println("myParam is", v)
}

// Ищем значение и отдаем
func get(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	fmt.Println(name)

	// Чтобы было только название очереди
	if r.RequestURI != r.URL.Path {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	answer := getFromQueue(name)
	if answer == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	// Передаем в теле ответ
	_, err := w.Write([]byte(answer))
	if err != nil {
		log.Println(err)
		return
	}

}

func main() {
	cont = make(map[string]*list.List)
	http.HandleFunc("/", Choice)
	fmt.Println("starting server at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
