package utils

import (
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

var (
	configPath string = "./config.yaml"
)

type FloodControl struct {
	MaxTries int `yaml:"maxTries"` // K - количество вызовов
	Interval int `yaml:"interval"` // N - секунд
	Counters map[int]int
	Ticker   *time.Ticker
}

func NewFloodControl() (*FloodControl, error) {
	fc := FloodControl{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data := yaml.NewDecoder(file)

	if err := data.Decode(&fc); err != nil {
		return nil, err
	}
	fc.Ticker = time.NewTicker(time.Second)
	fc.Counters = make(map[int]int)
	return &fc, nil

}
