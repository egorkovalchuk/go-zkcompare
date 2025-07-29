package data

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
)

// map c mutex
// для контроля потока записи. Мутекс для избегания блокировок
type Counters struct {
	mx sync.Mutex
	m  map[string]int
}

// Конструктор для типа данных Counters
func NewCounters() *Counters {
	return &Counters{
		m: make(map[string]int),
	}
}

// Получить значение
func (c *Counters) Load(key string) int {
	c.mx.Lock()
	val := c.m[key]
	c.mx.Unlock()
	return val
}

// Загрузить значение
func (c *Counters) Store(key string, value int) {
	c.mx.Lock()
	c.m[key] = value
	c.mx.Unlock()
}

// Инкримент +1
func (c *Counters) Inc(key string) {
	c.mx.Lock()
	c.m[key]++
	c.mx.Unlock()
}

// Загрузка в лог через функцию
func (c *Counters) LoadRangeToLogFunc(s string, f func(logtext interface{})) {
	c.mx.Lock()
	for k, v := range c.m {
		f(s + k + ": " + strconv.Itoa(v))
	}
	c.mx.Unlock()
}

// Загрузка
func (c *Counters) LoadRange() map[string]int {
	return c.m
}

// Config представляет корневую структуру конфигурации
type Config struct {
	Instances []Instance `json:"instances"`
}

// Instance описывает отдельный экземпляр конфигурации
type Instance struct {
	Source      string   `json:"source"`       // Источник данных
	Targets     []string `json:"targets"`      // Список целевых серверов
	Tags        []string `json:"tags"`         // Включаемые теги (пустой массив в примере)
	ExcludeTags []string `json:"exclude_tags"` // Исключаемые теги
	Path        string   `json:"path"`         // Путь к конфигурации
}

// Функция для загрузки конфигурации из файла
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Обрезка строки
func Cut(w []byte) string {
	if len(w) > 20 {
		return string(w[:20])
	} else {
		return string(w)
	}
}

type zkLoggerAdapter struct {
	logger *log.Logger
}

func (z *zkLoggerAdapter) Printf(format string, args ...interface{}) {
	z.logger.Printf(format, args)
}
