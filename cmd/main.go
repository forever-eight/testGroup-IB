package main

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var mainMap Map

func main() {
	portArg := flag.String("port", "8080", "http server port")
	flag.Parse()
	port := *portArg

	mainMap = Map{
		mutex: sync.Mutex{},
		m:     map[string]*Queue{},
	}

	http.HandleFunc("/", Choice)
	fmt.Println("starting server at :" + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

type Map struct {
	mutex sync.Mutex
	m     map[string]*Queue // Основная мапка
}

// Возвращает структуру по очереди
func (m *Map) getQ(name string) *Queue {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	q, ok := mainMap.m[name]
	if !ok {
		q = &Queue{
			answers: list.New(),
			waiters: list.New(),
		}
		mainMap.m[name] = q
	}
	return q
}

type Queue struct {
	m sync.Mutex

	answers *list.List // Ответы
	waiters *list.List // Каналы, которые ждут
}

// Добавление в очередь ответов
func (q *Queue) AddAnswers(param string) {
	q.answers.PushBack(param)
}

// Добавление в очередь каналов для связи с ждущими
func (q *Queue) AddWaiters(ch chan string) {
	q.waiters.PushBack(ch)
}

// Изъятие из очереди ответов
func (q *Queue) getFromAnswers() string {
	if q.answers.Len() > 0 {
		e := q.answers.Front() // Первый элемент
		defer q.answers.Remove(e)

		val := e.Value.(string)
		return val
	}

	return ""
}

// Изъятие из очереди каналов для ждущих
func (q *Queue) getFromWaiters() chan string {
	if q.waiters.Len() > 0 {
		e := q.waiters.Front() // Первый элемент
		defer q.waiters.Remove(e)

		val := e.Value.(chan string)
		return val
	}

	return nil
}

// Получаем наше значение и добавляем в очередь
func put(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	v := r.URL.Query().Get("v")
	if v == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	// Узнаем какая структура у нас конкретно для данной очереди
	q := mainMap.getQ(name)
	q.m.Lock()
	defer q.m.Unlock()

	// Берем первого ждущего из очереди и посылаем по каналу ему сразу правильный ответ
	ch := q.getFromWaiters()
	if ch != nil {
		ch <- v
		return
	}

	// Если у нас нет ждущего - добавляем в очередь
	q.AddAnswers(v)
}

func get(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	// Узнаем какая структура у нас конкретно для данной очереди и запрашиваем туда
	q := mainMap.getQ(name)

	q.m.Lock()

	answer := q.getFromAnswers()
	// Если timeout и ответа нет
	if len(r.URL.Path) < len(r.RequestURI) && answer == "" {
		ch := make(chan string)
		q.AddWaiters(ch)

		// Приведение типов
		N, err := strconv.Atoi(r.URL.Query().Get("timeout"))
		if err != nil {
			log.Println(err)
			return
		}
		sec := time.Duration(N)
		q.m.Unlock()
		timeout(w, ch, sec)

		return
	} else if answer == "" {
		q.m.Unlock()
		http.Error(w, "", http.StatusNotFound)
		return
	}
	// Передаем в теле ответ
	_, err := w.Write([]byte(answer))
	if err != nil {
		log.Println(err)
		return
	}
	q.m.Unlock()
}

// Выбор в случае, если нужно ждать
func timeout(w http.ResponseWriter, ch chan string, N time.Duration) {
	select {
	// Если пришло раньше из потока, нежели закончился таймер
	case news := <-ch:
		// Отвечаем
		_, err := w.Write([]byte(news))
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
