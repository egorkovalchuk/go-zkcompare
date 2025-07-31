package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/egorkovalchuk/go-zkcompare/data"
)

// Power by  Egor Kovalchuk

const (
	versionutil = "0.0.3"
	configname  = "config.json"
	// логи
	logFileName = "compare.log"
)

var (
	// запрос помощи
	help bool
	// запрос версии
	version bool
	// Дебаг режим
	debugm bool
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

	// Старт пути поиска/сравнения
	var pathzk string
	// Источник zk
	var sourcezk string
	// c чем сравниваем zk
	var dstzk string
	// Строка поиска
	var find string
	// Исключение
	var excl string
	// тип поиска
	var only string
	// Вывод пропущенных значений
	var printskeep bool
	// Включение Wathcher ZK, режим получение уведомлений об изменениях
	// работает только на выбранном путе, в дочернихне смотрит
	// или вешать на все
	var watcheron bool
	// Старт работы по конфигу. Сравнивает параметры между мастером и остальными
	var auto bool

	flag.StringVar(&sourcezk, "s", "", "Source Zookeeper is not empty")
	flag.StringVar(&dstzk, "d", "", "Destination Zookeeper is not empty")
	flag.StringVar(&pathzk, "p", "/", "Path Zookeeper, default /")
	flag.BoolVar(&debugm, "debug", false, "Debug mode")
	flag.StringVar(&excl, "e", "password", "exlude tags")
	flag.StringVar(&only, "only", "", "only empty, find only empty values")
	flag.StringVar(&find, "f", "", "find string")
	flag.BoolVar(&watcheron, "w", false, "Watcher zk mode")
	flag.BoolVar(&printskeep, "printskeep", false, "Print skeep values")
	flag.BoolVar(&auto, "auto", false, "Start application with config")
	flag.Parse()

	// запуск горутины записи в лог
	log := data.NewLogWriter(logFileName, debugm)
	go log.LogWriteForGoRutineStruct()

	log.ProcessInfo("- - - - - - - - - - - - - - -")
	log.ProcessInfo("Start report")

	switch {
	case find != "" && sourcezk != "":
		f := data.NewFind(sourcezk, pathzk, find, log)
		f.FindStart()
		return
	case watcheron && sourcezk != "":
		w := data.NewWatcher(sourcezk, pathzk, log)
		w.WatcherStart()
		return
	case auto:
		a, err := data.NewAuto("config.json", log)
		if err != nil {
			log.ProcessError(err)
			return
		} else {
			a.Start()
			return
		}
	default:
		excla := strings.Split(strings.ToLower(excl), ",")
		c := data.NewCompare(sourcezk, dstzk, pathzk, log, excla, only, printskeep, nil)
		c.CompareStart()
		return
	}
}

// Аналог Sleep.
func sleep(d time.Duration) {
	<-time.After(d)
}

func Helpstart() {
	fmt.Println("Start utill")
	fmt.Println("go-zkcompare -s source_zk -d dest_zk -p start_path -e excludetag1,excludetag2")
	fmt.Println("-s : Source Zookeeper address (mandatory parameter).")
	fmt.Println("-d : Destination Zookeeper address (mandatory parameter for compare).")
	fmt.Println("-p : Path in Zookeeper for comparison or search (default: /).")
	fmt.Println("-e : Excluding paths (comma-separated, default: password).")
	fmt.Println("-f : Search string (if specified, launches search mode).")
	fmt.Println("-debug : Enable debug mode (outputs additional information).")
	fmt.Println("-h : Display help.")
	fmt.Println("-v : Display the utility version.")
}
