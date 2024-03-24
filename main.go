package main

import (
	"context"
	"fmt"
	"log"
	"task/utils"
)

func main() {
	fc, err := utils.NewFloodControl()
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(fc)
}

// FloodControl интерфейс, который нужно реализовать.
// Рекомендуем создать директорию-пакет, в которой будет находиться реализация.
type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}
