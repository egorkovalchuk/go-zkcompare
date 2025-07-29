package data

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
)

type WatcherStruct struct {
	sourcezk string
	pathzk   string
	logFunc  func(string, interface{})
	srczkcon *zk.Conn
	// Контроль горутин
	wg sync.WaitGroup
}

func NewWatcher(sourcezk string, pthzk string, logFunc func(string, interface{})) *WatcherStruct {
	return &WatcherStruct{
		sourcezk: sourcezk,
		pathzk:   pthzk,
		logFunc:  logFunc,
	}
}

func (w *WatcherStruct) WatcherStart() {
	w.logFunc("INFO", "Start watcher zk")
	var err error
	w.srczkcon, _, err = zk.Connect([]string{w.sourcezk}, time.Second*10)

	if err != nil {
		w.logFunc("PANIC", err)
	}

	// Канал для получения событий
	eventChan := make(chan zk.Event)
	// Устанавливаем watcher
	_, _, watcher, err := w.srczkcon.ChildrenW(w.pathzk)
	if err != nil {
		panic(err)
	}
	// Горутина для обработки событий
	go func() {
		for {
			select {
			case event := <-watcher:
				eventChan <- event
				// Переустанавливаем watcher после получения события
				_, _, watcher, err = w.srczkcon.ChildrenW(w.pathzk)
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	// Обработчик событий
	for {
		event := <-eventChan
		fmt.Printf("Событие: %+v\n", event)

		switch event.Type {
		case zk.EventNodeCreated:
			fmt.Println("Узел создан")
		case zk.EventNodeDeleted:
			fmt.Println("Узел удален")
		case zk.EventNodeDataChanged:
			data, _, err := w.srczkcon.Get(w.pathzk)
			if err != nil {
				fmt.Println("Ошибка получения данных:", err)
				continue
			}
			fmt.Printf("Данные изменены: %s\n", data)
		case zk.EventNodeChildrenChanged:
			fmt.Println("Дочерние узлы изменены")
			fmt.Println(event.Path)
		}
	}

}
