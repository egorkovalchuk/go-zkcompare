package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
)

// Power by  Egor Kovalchuk

const (
	// логи
	logFileName = "compare.log"
	versionutil = "0.0.2"
)

var (
	// Запись в лог
	filer *os.File
	// Запись в лог
	rezfiler *os.File
	// запрос помощи
	help bool
	// ошибки
	err error
	// запрос версии
	version bool
	// Источник zk
	sourcezk string
	// c чем сравниваем zk
	dstzk string
	// исключение
	excl  string
	excla []string
	// старт пути
	pathzk    string
	srczkcon  *zk.Conn
	destzkcon *zk.Conn

	// Строка поиска
	find string
	// Дебаг режим
	debugm bool

	wg sync.WaitGroup
)

func main() {
	//start program
	var argument string
	if len(os.Args) > 1 {
		argument = os.Args[1]
	} else {
		Helpstart()
		return
	}

	if argument == "-h" {
		Helpstart()
		return
	} else if argument == "-v" {
		fmt.Println("Version utill " + versionutil)
		return
	}

	flag.StringVar(&sourcezk, "s", "", "Source Zookeeper is not empty")
	flag.StringVar(&dstzk, "d", "", "Destination Zookeeper is not empty")
	flag.StringVar(&pathzk, "p", "/", "Path Zookeeper, default /")
	flag.BoolVar(&debugm, "debug", false, "Debug mode")
	flag.StringVar(&excl, "e", "password", "exlude tags")
	flag.StringVar(&find, "f", "", "find string")
	flag.Parse()

	// Открытие лог файла
	// ротация не поддерживается в текущей версии
	// Вынести в горутину
	filer, err := os.OpenFile(logFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer filer.Close()

	log.SetOutput(filer)

	// запуск горутины записи в лог
	go LogWriteForGoRutineStruct(LogChannel)

	ProcessInfo("- - - - - - - - - - - - - - -")
	ProcessInfo("Start report")

	if find != "" && sourcezk != "" {
		FindStart()
		return
	} else {
		CompareStart()
		return
	}
}

func ReChildren(chdl []string, pthzk string) {
	defer wg.Done()
	for _, i := range chdl {
		if !CompareZk(i) {
			tmp := pthzk + "/" + i
			ProcessDebug("Check " + tmp)

			sgg, _, err := srczkcon.Get(tmp)
			if len(sgg) > 1 {
				ProcessDebug(string(sgg))
			}

			dgg, _, err := destzkcon.Get(tmp)

			if !bytes.Equal(sgg, dgg) {
				if len(sgg) > 30 && len(dgg) > 30 {
					ProcessWarm("source :" + string(tmp) + " value: ***(big value) ")
				} else {
					ProcessWarm("source :" + string(tmp) + " value: " + Cut(sgg) + " value destination: " + Cut(dgg))
				}
			}

			//Идем дальше
			children, _, err := srczkcon.Children(tmp)
			if err != nil {
				ProcessError(err)
			} else {
				wg.Add(1)
				go ReChildren(children, tmp)
			}
		} else {
			ProcessInfo("Skeep: " + i)
		}
	}
}

// Исключение полей из строки запуска
func CompareZk(pth string) bool {
	check := false
	for _, i := range excla {
		if strings.Contains(pth, i) {
			return true
		}
	}
	return check
}

func CompareStart() {
	if sourcezk == "" || pathzk == "" || dstzk == "" {
		fmt.Println("Use go-zkcompare -s source_zk -d dest_zk -p start_path")
		return
	}

	excla = strings.Split(excl, ",")

	srczkcon, _, err = zk.Connect([]string{sourcezk}, time.Second*10)
	if err != nil {
		ProcessPanic(err)
	}

	destzkcon, _, err = zk.Connect([]string{dstzk}, time.Second*10)
	if err != nil {
		ProcessPanic(err)
	}

	children, _, err := srczkcon.Children(pathzk)
	if err != nil {
		ProcessPanic(err)
	} else {
		wg.Add(1)
		go ReChildren(children, pathzk)
	}

	wg.Wait() // Ожидаем завершения всех горутин
	ProcessInfo("Stop find")
	sleep(2 * time.Second)
}

func FindStart() {
	ProcessInfo("Start find")
	srczkcon, _, err = zk.Connect([]string{sourcezk}, time.Second*10)

	if err != nil {
		ProcessPanic(err)
	}

	children, _, err := srczkcon.Children(pathzk)

	if err != nil {
		ProcessPanic(err)
	} else {
		wg.Add(1)
		go ReChildrenFind(children, pathzk)
	}

	wg.Wait() // Ожидаем завершения всех горутин
	ProcessInfo("Stop find")
	sleep(2 * time.Second)
}

func ReChildrenFind(chdl []string, pthzk string) {
	defer wg.Done()
	for _, i := range chdl {

		tmp := pthzk + "/" + i

		sgg, _, err := srczkcon.Get(tmp)
		if len(sgg) > 1 {
			ProcessDebug(string(tmp))
		}

		if strings.Contains(string(sgg), find) {
			ProcessInfo("source :" + string(tmp) + " value: " + Cut(sgg))
		}

		//Идем дальше
		children, _, err := srczkcon.Children(tmp)
		if err != nil {
			ProcessError(err)
		} else {
			wg.Add(1)
			go ReChildrenFind(children, tmp)
		}

	}

}

// Обрезка строки
func Cut(w []byte) string {
	if len(w) > 20 {
		return string(w[:20])
	} else {
		return string(w)
	}
}
