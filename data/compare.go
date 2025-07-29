package data

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
)

type CompareStruct struct {
	sourcezk string
	pathzk   string
	dstzk    string
	logFunc  func(string, interface{})
	excla    []string
	include  []string
	srczkcon *zk.Conn
	dstzkcon *zk.Conn
	// Счетчик для подсчеа позиций сравнения
	zkcount *Counters
	// Контроль горутин
	wg sync.WaitGroup
	// Ограничение поиска
	// на пример empty. только поиск пустых
	tag string
	// вывод равных параметров
	printskeep bool
	logger     *log.Logger
}

func NewCompare(sourcezk string, dstzk string, pthzk string, logFunc func(string, interface{}), excla []string, tag string, printskeep bool, logger *log.Logger, include []string) *CompareStruct {
	return &CompareStruct{
		sourcezk:   sourcezk,
		pathzk:     pthzk,
		dstzk:      dstzk,
		logFunc:    logFunc,
		excla:      excla,
		tag:        tag,
		printskeep: printskeep,
		zkcount:    NewCounters(),
		logger:     logger,
		include:    include,
	}
}

func (c *CompareStruct) CompareStart() {
	if c.sourcezk == "" || c.pathzk == "" || c.dstzk == "" {
		fmt.Println("Use go-zkcompare -s source_zk -d dest_zk -p start_path")
		return
	}

	var err error
	zkLogger := zkLoggerAdapter{logger: c.logger}
	c.srczkcon, _, err = zk.Connect([]string{c.sourcezk}, time.Second*10, zk.WithLogger(&zkLogger))
	if err != nil {
		c.logFunc("PANIC", err)
	}
	c.dstzkcon, _, err = zk.Connect([]string{c.dstzk}, time.Second*10, zk.WithLogger(&zkLogger))
	if err != nil {
		c.logFunc("PANIC", err)
	}

	// Добавить %

	children, _, err := c.srczkcon.Children(c.pathzk)
	if err != nil {
		c.logFunc("PANIC", err)
	} else {
		c.wg.Add(1)
		go c.ReChildren(children, c.pathzk)
	}

	c.wg.Wait() // Ожидаем завершения всех горутин
	c.logFunc("INFO", "Stop compare")
	for k, v := range c.zkcount.LoadRange() {
		c.ProcessInfo(k + ": " + strconv.Itoa(v))
	}
	sleep(2 * time.Second)
}

func (c *CompareStruct) ReChildren(chdl []string, pthzk string) {
	defer c.wg.Done()
	for _, i := range chdl {
		if !c.CompareZk(i) {
			tmp := pthzk + "/" + i
			c.logFunc("DEBUG", "Check "+tmp)

			sgg, _, err := c.srczkcon.Get(tmp)
			if len(sgg) > 1 {
				c.logFunc("DEBUG", string(sgg))
			}

			dgg, _, err := c.dstzkcon.Get(tmp)

			if !bytes.Equal(sgg, dgg) {
				switch {
				case c.tag == "empty":
					if len(dgg) == 0 {
						c.logFunc("WARM", "source :"+string(tmp)+" value destination empty")
						c.zkcount.Inc("EMPTY")
					}
				case len(sgg) > 30 && len(dgg) > 30:
					c.logFunc("WARM", "source :"+string(tmp)+" value: ***(big value) ")
					c.zkcount.Inc("UNEQUAL BIG")
				case len(dgg) == 0:
					c.logFunc("WARM", "source :"+string(tmp)+" value destination empty")
					c.zkcount.Inc("EMPTY")
				default:
					c.logFunc("WARM", "source :"+string(tmp)+" value: "+Cut(sgg)+" value destination: "+Cut(dgg))
					c.zkcount.Inc("UNEQUAL")
				}
			} else {
				c.zkcount.Inc("QUAL")
			}

			//Идем дальше
			children, _, err := c.srczkcon.Children(tmp)
			if err != nil {
				c.logFunc("ERROR", err)
			} else {
				c.wg.Add(1)
				go c.ReChildren(children, tmp)
			}
		} else {
			if c.printskeep {
				c.logFunc("INFO", "Skeep: "+pthzk+"/"+i)
			}
			c.zkcount.Inc("SKEEP")
		}
	}
}

// Исключение полей из строки запуска
func (c *CompareStruct) CompareZk(pth string) bool {
	check := false
	for _, i := range c.excla {
		//if strings.Contains(strings.ToLower(pth), i) {
		if strings.ToLower(pth) == i {
			return true
		}
	}
	return check
}

func (c *CompareStruct) ProcessInfo(logtext interface{}) {
	c.logFunc("INFO", logtext)
}
