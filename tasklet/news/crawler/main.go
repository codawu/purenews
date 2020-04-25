package main

import (
	"fmt"
	"os"
	"purenews/prepare"
	"time"
)

type NewsTask struct {
	Retry int
	Name  string
}

func (task *NewsTask) Run() (interface{}, error) {
	return nil, nil
}
func (task *NewsTask) SetRetry(retry int) {
	task.Retry = retry
}
func (task *NewsTask) GetRetry() int {
	return task.Retry
}

type NewsTaskQueue struct {
}

func (queue *NewsTaskQueue) Pop() (prepare.Task, error) {
	return new(NewsTask), nil
}

func (queue *NewsTaskQueue) Push(prepare.Task) error {
	return nil
}

func main() {
	tick := time.Tick(time.Second / time.Duration(500))
	if err := prepare.Worker(&tick, new(NewsTaskQueue)); err != nil {
		fmt.Printf("worker start error %q\n", err)
		os.Exit(-1)
	}
}
