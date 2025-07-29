package data

import (
	"log"
	"sync"
	"time"
)

type CompareAuto struct {
	cfg Config
	// Счетчик для подсчеа позиций сравнения
	zkcount *Counters
	// Контроль горутин
	wg      sync.WaitGroup
	logFunc func(string, interface{})
	logger  *log.Logger
}

func NewAuto(confname string, logFunc func(string, interface{}), logger *log.Logger) (*CompareAuto, error) {
	cfgt, err := LoadConfig(confname)
	if err != nil {
		return nil, err
	}
	return &CompareAuto{cfg: *cfgt, logFunc: logFunc, logger: logger}, nil
}

func (c *CompareAuto) Start() {
	//	var err error
	for _, i := range c.cfg.Instances {
		for _, j := range i.Targets {
			c.logFunc("INFO", j)
			com := NewCompare(i.Source, j, i.Path, c.logFunc, i.ExcludeTags, "", false, c.logger, i.Tags)
			com.CompareStart()
		}
	}
	sleep(1 * time.Second)
}
