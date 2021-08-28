package main

import (
	"container/list"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

//todo подумать над структурой для ответа ждуняшек в очереди
var cont map[string]*list.List

//todo структура с каналом и листом
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
		return val
	}

	return ""
}

func timeout(w http.ResponseWriter, r *http.Request, n string, quit chan int) {
	N, err := strconv.Atoi(n)
	if err != nil {
		log.Println(err)
		return
	}

	// todo сделать там приведение к time.Duration
	// todo ретерним строку с ответом и по каналу передаем готов ли у нас
	select {
	case news := <-quit:
		fmt.Println(news)
	case <-time.After(time.Duration(N) * time.Second):
		fmt.Println("No news in five seconds.")
	}

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
	// todo  если пришло то, что мы ждем в гете - отправляем его сразу туда

	name := r.URL.Path[1:]
	v := r.URL.Query().Get("v")
	addToQueue(name, v)
	if v == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

}

// Распределяем get (на с timeout и без )
func get(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	answer := getFromQueue(name) // answer, ok
	//todo get from queue сделать второе окей на пустоту проверка
	if len(r.URL.Path) < len(r.RequestURI) && answer == "" {
		quit := make(chan int)
		timeout(w, r, r.URL.Query().Get("timeout"), quit)
		return
	} else if answer == "" {
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
