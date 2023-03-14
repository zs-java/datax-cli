package libdatax

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestThreadPool(t *testing.T) {

	pl := NewFixedSizeThreadPool(3)

	var tasks []Task
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("Task-%d", i)
		task := Task{
			Name: name,
			Action: func() {
				log.Println("begin ", name)
				time.Sleep(time.Second * 2)
				log.Println("end ", name)
			},
		}
		tasks = append(tasks, task)
		pl.SubmitTask(task)
	}

	pl.Wait()

}
