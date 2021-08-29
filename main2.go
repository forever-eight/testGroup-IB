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

//todo структура с каналом и листом

var q Queue
var countWaiters int = 0

type Queue struct {
	cont     map[string]*list.List // контейнер с названием очереди
	channels []chan string         // каналы, по которым у нас передаются данные
	waiters  map[string]int        // очередь из запросов
}

// Добавление в очередь
func addToQueue(name string, param string) {
	if q.cont[name] == nil {
		queue := list.New()
		queue.PushBack(param)
		q.cont[name] = queue
	} else {
		q.cont[name].PushBack(param)
	}

}

// Изъятие из очереди
func getFromQueue(name string) string {
	answer, ok := q.cont[name]
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

func timeout(w http.ResponseWriter, r *http.Request, N time.Duration) {

	// todo сделать там приведение к time.Duration
	// todo ретерним строку с ответом и по каналу передаем готов ли у нас
	select {
	// Если пришло раньше из потока, нежели закончился таймер
	case news := <-q.channels[countWaiters]:
		answer := getFromQueue(news)
		// Отвечаем
		_, err := w.Write([]byte(answer))
		if err != nil {
			log.Println(err)
			return
		}
	case <-time.After(N * time.Second):
		http.Error(w, "", http.StatusNotFound)
		return
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
	// Ловим данные из канала
	needed := <-q.channels[countWaiters]

	name := r.URL.Path[1:]
	v := r.URL.Query().Get("v")
	if v == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	addToQueue(name, v)

	// Если кто-то ждет и наше значение равно тому, что человек ждет
	if needed != "" && needed == name {
		q.channels[countWaiters] <- name
	}

}

// Распределяем get (на с timeout и без )
func get(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	answer := getFromQueue(name) // answer, ok
	//todo get from queue сделать второе окей на пустоту проверка
	if len(r.URL.Path) < len(r.RequestURI) && answer == "" {

		// Прибавляем к количеству ждущих или говорим, что их теперь 1 человек
		_, ok := q.waiters[name]
		if !ok {
			q.waiters[name] = 1
		} else {
			q.waiters[name]++
		}
		countWaiters++
		// Создаем канал для передачи данных
		quit := make(chan string, 5)
		q.channels[countWaiters] = quit
		q.channels[countWaiters] <- name
		// Приведение типов
		N, err := strconv.Atoi(r.URL.Query().Get("timeout"))
		if err != nil {
			log.Println(err)
			return
		}
		sec := time.Duration(N)
		timeout(w, r, sec)

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
	cont := make(map[string]*list.List)
	wait := make(map[string]int)
	q = Queue{
		cont:    cont,
		waiters: wait,
	}

	http.HandleFunc("/", Choice)
	fmt.Println("starting server at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
