package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Etozheraf/TaskVK/pkg/workers"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL)
	defer cancel()

	fmt.Println("Введите количество воркеров")
	var n int
	_, err := fmt.Scan(&n)
	if err != nil {
		return
	}

	jobs := make(chan string)
	workerPool := workers.NewWorkerPool(ctx, jobs)
	for i := range n {
		workerPool.AddWorker(i, func(s string) {
			fmt.Printf("Worker id: %d Message: %s\n", i, s)
		})
	}

	fmt.Println("Введите сообщения (q для выхода)")
	var str string
	for {
		_, err := fmt.Scan(&str)
		if err != nil {
			return
		}
		if str == "q" {
			break
		}

		jobs <- str
	}

	workerPool.Shutdown()
	close(jobs)
}
