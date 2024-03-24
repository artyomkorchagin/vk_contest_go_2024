package utils

import (
	"context"
	"gopkg.in/yaml.v2"
	"os"
	"sync"
	"time"
)

var (
	configPath = "./config.yaml"
)

type FloodControl struct {
	MaxTries int `yaml:"maxTries"` // K - количество вызовов
	Interval int `yaml:"interval"` // N - секунд
	Mutex    sync.Mutex
	Counters map[int64]int
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
	fc.Ticker = time.NewTicker(time.Second * time.Duration(fc.Interval))
	fc.Counters = make(map[int64]int)
	return &fc, nil
}

func (f *FloodControl) Check(ctx context.Context, userID int64) (bool, error) {
	f.Counters[userID]++
	if f.Counters[userID] > f.MaxTries {
		return false, nil
	}

	go func() {
		for range f.Ticker.C {
			f.Mutex.Lock()
			for userID := range f.Counters {
				f.Counters[userID] = 0
			}
			f.Mutex.Unlock()
		}
	}()
	return true, nil

}
