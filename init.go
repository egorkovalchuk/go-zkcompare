package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	configname = "config.json"
	// логи
	logFileName = "compare.log"
)

type LogStruct struct {
	t    string
	text interface{}
}

var (
	LogChannel = make(chan LogStruct)
	// Запись в лог
	filer  *os.File
	logger *log.Logger
)

// Запись ошибок из горутин
// можно добавить ротейт по дате + архив в отдельном потоке
func LogWriteForGoRutineStruct(logs chan LogStruct) {
	filer, err := os.OpenFile(logFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer filer.Close()
	logger = log.New(filer, "", 0)

	for entry := range logs {
		prefix := time.Now().Local().Format("2006/01/02 15:04:05") + " " + entry.t + ": "
		logger.SetPrefix(prefix)
		logger.Println(entry.text)
	}
}

// Запись в лог при включенном дебаге
func ProcessDebug(logtext interface{}) {
	if debugm {
		LogChannel <- LogStruct{"DEBUG", logtext}
	}
}

// Запись в лог ошибок
func ProcessError(logtext interface{}) {
	LogChannel <- LogStruct{"ERROR", logtext}
}

// Запись в лог ошибок cсо множеством переменных
func ProcessErrorAny(logtext ...interface{}) {
	t := ""
	for _, a := range logtext {
		t += fmt.Sprint(a) + " "
	}
	LogChannel <- LogStruct{"ERROR", t}
}

// Запись в лог WARM
func ProcessWarm(logtext interface{}) {
	LogChannel <- LogStruct{"WARM", logtext}
}

// Запись в лог INFO
func ProcessInfo(logtext interface{}) {
	LogChannel <- LogStruct{"INFO", logtext}
}

// Запись в лог Diam
func ProcessDiam(logtext interface{}) {
	LogChannel <- LogStruct{"DIAM", logtext}
}

// Запись в лог Influx
func ProcessInflux(logtext interface{}) {
	LogChannel <- LogStruct{"INFLUX", logtext}
}

// Нештатное завершение при критичной ошибке
func ProcessPanic(logtext interface{}) {
	fmt.Println(logtext)
	os.Exit(2)
}

// Запись в лог
func ProcessLog(level string, logtext interface{}) {
	switch level {
	case "DEBUG":
		ProcessDebug(logtext)
	case "PANIC":
		ProcessPanic(logtext)
	default:
		LogChannel <- LogStruct{level, logtext}
	}
}

// Инициализация переменных
func InitVariables() {}

// Аналог Sleep.
func sleep(d time.Duration) {
	<-time.After(d)
}

func Helpstart() {
	fmt.Println("Start utill")
	fmt.Println("go-zkcompare -s source_zk -d dest_zk -p start_path -e excludetag1,excludetag2")
	fmt.Println("Use -v get version")
}
