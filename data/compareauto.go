package data

import (
	"sync"
	"time"
)

type CompareAuto struct {
	cfg Config
	// Счетчик для подсчеа позиций сравнения
	zkcount *Counters
	// Контроль горутин
	wg      sync.WaitGroup
	logFunc *LogWriter
}

func NewAuto(confname string, logFunc *LogWriter) (*CompareAuto, error) {
	cfgt, err := LoadConfig(confname)
	if err != nil {
		return nil, err
	}
	return &CompareAuto{cfg: *cfgt, logFunc: logFunc}, nil
}

func (c *CompareAuto) Start() {
	//	var err error
	for _, i := range c.cfg.Instances {
		for _, j := range i.Targets {
			c.logFunc.ProcessInfo(j)
			com := NewCompare(i.Source, j, i.Path, c.logFunc, i.ExcludeTags, "", true, i.Tags)
			com.CompareStart()
		}
	}
	sleep(1 * time.Second)
}
