package gogular

import (
	"os"
	"strconv"
	"fmt"
)

/**
	For optimization performance analysis
 */

type Analysis struct {
	StartSize   int64
	EndSize     int64

	Improvement float64

	Tasks       []Task
}

type Task struct {
	Task        string

	BeforeSize  int64
	AfterSize   int64

	Improvement float64
}

func (a *Analysis) LogSize(name string, f1 *os.File, f2 *os.File) {
	task := Task{
		Task: name,
	}

	f1s, err := f1.Stat()
	if err != nil {
		fmt.Print(err)
	}
	f2s, err := f2.Stat()
	if err != nil {
		fmt.Print(err)
	}

	if f1s != nil {
		task.BeforeSize = f1s.Size()
	}

	if f2s != nil {
		task.AfterSize = f2s.Size()
	}

	task.Improvement = (float64(task.BeforeSize) - float64(task.AfterSize)) / float64(task.BeforeSize)

	if len(a.Tasks) == 0 {
		a.Tasks = []Task{}
	}

	a.Tasks = append(a.Tasks, task)

	a.StartSize += task.BeforeSize
	a.EndSize += task.AfterSize

	a.Improvement = (float64(a.StartSize) - float64(a.EndSize)) / float64(a.StartSize)
}

func (a *Analysis) String() {
	fmt.Print("Analysis")
	for _, t := range a.Tasks {
		impr := strconv.FormatFloat(t.Improvement, 'f', 6, 32)
		fmt.Print("Task: " + t.Task + ", Size improvement: " + impr)
	}
}