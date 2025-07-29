package data

import (
	"strings"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
)

type FingStruct struct {
	sourcezk string
	pathzk   string
	find     string
	logFunc  func(string, interface{})
	srczkcon *zk.Conn
	// Контроль горутин
	wg sync.WaitGroup
}

func NewFind(sourcezk string, pthzk string, find string, logFunc func(string, interface{})) *FingStruct {
	return &FingStruct{
		sourcezk: sourcezk,
		pathzk:   pthzk,
		find:     find,
		logFunc:  logFunc,
	}
}

func (f *FingStruct) FindStart() {
	f.logFunc("INFO", "Start find")
	var err error
	f.srczkcon, _, err = zk.Connect([]string{f.sourcezk}, time.Second*10)

	if err != nil {
		f.logFunc("ERROR", err)
	}

	children, _, err := f.srczkcon.Children(f.pathzk)

	if err != nil {
		f.logFunc("PANIC", err)
	} else {
		f.wg.Add(1)
		go f.ReChildrenFind(children, f.pathzk)
	}

	f.wg.Wait() // Ожидаем завершения всех горутин
	f.logFunc("INFO", "Stop find")
	sleep(2 * time.Second)
}

func (f *FingStruct) ReChildrenFind(chdl []string, pthzk string) {
	defer f.wg.Done()
	for _, i := range chdl {

		tmp := pthzk + "/" + i

		sgg, _, err := f.srczkcon.Get(tmp)
		if len(sgg) > 1 {
			f.logFunc("DEBUG", string(tmp))
		}

		if strings.Contains(string(sgg), f.find) {
			f.logFunc("INFO", "source :"+string(tmp)+" value: "+Cut(sgg))
		}

		//Идем дальше
		children, _, err := f.srczkcon.Children(tmp)
		if err != nil {
			f.logFunc("ERROR", err)
		} else {
			f.wg.Add(1)
			go f.ReChildrenFind(children, tmp)
		}

	}

}

// Аналог Sleep.
func sleep(d time.Duration) {
	<-time.After(d)
}
