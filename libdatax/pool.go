package libdatax

import (
	"sync"
)

type Task struct {
	Name   string
	Action func()
}

type ThreadPool interface {
	SubmitTask(task Task)
	Wait()
}

type FixedSizeThreadPool struct {
	runningTaskChan chan Task
	wg              sync.WaitGroup
}

func NewFixedSizeThreadPool(size int) *FixedSizeThreadPool {
	return &FixedSizeThreadPool{
		runningTaskChan: make(chan Task, size),
	}
}

func (t *FixedSizeThreadPool) SubmitTask(task Task) {
	t.wg.Add(1)
	go func() {
		t.runningTaskChan <- task
		task.Action()
		t.wg.Done()
		<-t.runningTaskChan
	}()
}

func (t *FixedSizeThreadPool) Wait() {
	t.wg.Wait()
}
