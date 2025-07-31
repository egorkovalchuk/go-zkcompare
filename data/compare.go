package data

import (
	"bytes"
	"fmt"
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
	logFunc  *LogWriter
	// список исключений
	excla []string
	// список включений
	include  []string
	srczkcon *zk.Conn
	dstzkcon *zk.Conn
	// Счетчик для подсчеа позиций сравнения
	zkcount *Counters
	// Контроль горутин
	wg sync.WaitGroup
	// Ограничение поиска
	// на пример empty. только поиск пустых
	only string
	// вывод равных параметров
	printskeep bool
}

func NewCompare(sourcezk string, dstzk string, pthzk string, logFunc *LogWriter, excla []string, only string, printskeep bool, include []string) *CompareStruct {
	return &CompareStruct{
		sourcezk:   sourcezk,
		pathzk:     pthzk,
		dstzk:      dstzk,
		logFunc:    logFunc,
		excla:      excla,
		only:       only,
		printskeep: printskeep,
		zkcount:    NewCounters(),
		include:    include,
	}
}

func (c *CompareStruct) CompareStart() {
	if c.sourcezk == "" || c.pathzk == "" || c.dstzk == "" {
		fmt.Println("Use go-zkcompare -s source_zk -d dest_zk -p start_path")
		return
	}

	var err error
	zkLogger := zkLoggerAdapter{logger: c.logFunc.GetLogger()}
	c.srczkcon, _, err = zk.Connect([]string{c.sourcezk}, time.Second*10, zk.WithLogger(&zkLogger))
	if err != nil {
		c.logFunc.ProcessPanic(err)
	}
	c.dstzkcon, _, err = zk.Connect([]string{c.dstzk}, time.Second*10, zk.WithLogger(&zkLogger))
	if err != nil {
		c.logFunc.ProcessPanic(err)
	}

	// Добавить %

	children, _, err := c.srczkcon.Children(c.pathzk)
	if err != nil {
		c.logFunc.ProcessPanic(err)
	} else {
		c.wg.Add(1)
		go c.ReChildren(children, c.pathzk)
	}

	c.wg.Wait() // Ожидаем завершения всех горутин
	c.logFunc.ProcessInfo("Stop compare")
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
			c.logFunc.ProcessDebug("Check " + tmp)

			sgg, _, err := c.srczkcon.Get(tmp)
			if len(sgg) > 1 {
				c.logFunc.ProcessDebug(string(sgg))
			}

			dgg, _, err := c.dstzkcon.Get(tmp)

			if !bytes.Equal(sgg, dgg) {
				switch {
				case c.only == "empty":
					if len(dgg) == 0 {
						c.logFunc.ProcessWarm("source :" + string(tmp) + " value destination empty")
						c.zkcount.Inc("EMPTY")
					}
				case len(sgg) > 30 && len(dgg) > 30:
					c.logFunc.ProcessWarm("source :" + string(tmp) + " value: ***(big value) ")
					c.zkcount.Inc("UNEQUAL BIG")
				case len(dgg) == 0:
					c.logFunc.ProcessWarm("source :" + string(tmp) + " value destination empty")
					c.zkcount.Inc("EMPTY")
				default:
					c.logFunc.ProcessWarm("source :" + string(tmp) + " value: " + Cut(sgg) + " value destination: " + Cut(dgg))
					c.zkcount.Inc("UNEQUAL")
				}
			} else {
				c.zkcount.Inc("QUAL")
			}

			//Идем дальше
			children, _, err := c.srczkcon.Children(tmp)
			if err != nil {
				c.logFunc.ProcessError(err)
			} else {
				c.wg.Add(1)
				go c.ReChildren(children, tmp)
			}
		} else {
			if c.printskeep {
				c.logFunc.ProcessInfo("Skeep: " + pthzk + "/" + i)
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
	c.logFunc.ProcessInfo(logtext)
}
