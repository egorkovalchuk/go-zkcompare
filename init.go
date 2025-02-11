package main

import (
	"fmt"

	"log"
	"os"
	"time"
)

type LogStruct struct {
	t    string
	text interface{}
}

var (
	LogChannel = make(chan LogStruct)
)

// Запись ошибок из горутин
// можно добавить ротейт по дате + архив в отдельном потоке
func LogWriteForGoRutineStruct(err chan LogStruct) {
	for i := range err {
		datetime := time.Now().Local().Format("2006/01/02 15:04:05")
		log.SetPrefix(datetime + " " + i.t + ": ")
		log.SetFlags(0)
		log.Println(i.text)
		log.SetPrefix("")
		log.SetFlags(log.Ldate | log.Ltime)
	}
}

// Запись в лог при включенном дебаге
// Сделать горутиной?
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
