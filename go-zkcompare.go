package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

//Power by  Egor Kovalchuk

const (
	// логи
	logFileName     = "compare.log"
	versionutil     = "0.0.1"
	CompareFileName = "rez.log"
)

var (
	//Запись в лог
	filer *os.File
	//Запись в лог
	rezfiler *os.File
	//запрос помощи
	help bool
	//ошибки
	err error
	//запрос версии
	version bool
	//Источник zk
	sourcezk string
	//c чем сравниваем zk
	dstzk string
	//исключение
	excl  string
	excla []string
	//старт пути
	pathzk    string
	srczkcon  *zk.Conn
	destzkcon *zk.Conn
)

func main() {
	//start program
	filer, err = os.OpenFile(logFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(filer)
	log.Println("- - - - - - - - - - - - - - -")
	log.Println("Start report")

	flag.BoolVar(&version, "v", false, "a bool")
	flag.BoolVar(&help, "h", false, "a bool")
	flag.StringVar(&sourcezk, "s", "", "Source Zookeeper is not empty")
	flag.StringVar(&dstzk, "d", "", "Destination Zookeeper is not empty")
	flag.StringVar(&pathzk, "p", "", "Path Zookeeper is not empty")
	flag.StringVar(&excl, "e", "", "")
	flag.Parse()

	if version {
		fmt.Println("Version utill " + versionutil)
		return
	}

	if help {
		Helpstart()
		return
	}

	if sourcezk == "" || pathzk == "" || dstzk == "" {
		fmt.Println("Use go-zkcompare -s source_zk -d dest_zk -p start_path")
		return
	}

	excla = strings.Split(excl, ",")

	rezfiler, err = os.OpenFile(CompareFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		log.Fatal(err)
	}

	srczkcon, _, err = zk.Connect([]string{sourcezk}, time.Second) //*10)
	if err != nil {
		panic(err)
	}

	destzkcon, _, err = zk.Connect([]string{dstzk}, time.Second) //*10)
	if err != nil {
		panic(err)
	}

	children, _, err := srczkcon.Children(pathzk)
	if err != nil {
		panic(err)
	} else {
		ReChildren(children, pathzk)
	}

	defer rezfiler.Close()
	defer filer.Close()

}

func ReChildren(chdl []string, pthzk string) {
	for _, i := range chdl {
		if !CompareZk(i) {
			tmp := pthzk + "/" + i
			log.Println("Check " + tmp)

			sgg, _, err := srczkcon.Get(tmp)
			if len(sgg) > 1 {
				log.Println(string(sgg))
			}

			dgg, _, err := destzkcon.Get(tmp)

			if !bytes.Equal(sgg, dgg) {
				rezfiler.WriteString(tmp + "\n")
				rezfiler.WriteString("source :" + string(sgg) + " destination: " + string(dgg) + "\n")
			}

			//Идем дальше
			children, _, err := srczkcon.Children(tmp)
			if err != nil {
				log.Println(err)
			} else {
				ReChildren(children, tmp)
			}
		}
	}
}

func CompareZk(pth string) bool {
	check := false
	for _, i := range excla {
		if strings.Contains(pth, i) {
			return true
		}
	}
	return check
}

func Helpstart() {
	fmt.Println("Start utill")
	fmt.Println("go-zkcompare -s source_zk -d dest_zk -p start_path -e excludetag1,excludetag2")
	fmt.Println("Use -v get version")
}
