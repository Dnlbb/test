package main

import (
	"fmt"
	"sync"
	"time"
)

type (
	WorkerPool struct {
		tasks        chan string
		addWorker    chan struct{}
		removeWorker chan struct{}
		workers      []*Worker
		wg           sync.WaitGroup
	}

	Worker struct {
		id          int
		taskChannel <-chan string
		stop        chan struct{}
	}
)

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		tasks:        make(chan string),
		addWorker:    make(chan struct{}),
		removeWorker: make(chan struct{}),
	}
}

func (wp *WorkerPool) Start() {
	go func() {
		id := 1
		for {
			select {
			case <-wp.addWorker:
				worker := NewWorker(id, wp.tasks)
				wp.workers = append(wp.workers, worker)
				worker.Start(&wp.wg)
				id++
				fmt.Printf("add worker %d\n", worker.id)
			case <-wp.removeWorker:
				if len(wp.workers) == 0 {
					continue
				}
				worker := wp.workers[len(wp.workers)-1]
				wp.workers = wp.workers[:len(wp.workers)-1]
				worker.Stop()
				fmt.Printf("removed worker %d\n", worker.id)
			}
		}
	}()
}

func (wp *WorkerPool) AddWorker() {
	wp.addWorker <- struct{}{}
}

func (wp *WorkerPool) RemoveWorker() {
	wp.removeWorker <- struct{}{}
}

func (wp *WorkerPool) AddTask(task string) {
	wp.tasks <- task
}

func (wp *WorkerPool) Stop() {
	close(wp.tasks)
	wp.wg.Wait()
}

func NewWorker(id int, taskChannel <-chan string) *Worker {
	return &Worker{
		id:          id,
		taskChannel: taskChannel,
		stop:        make(chan struct{}),
	}
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case task, ok := <-w.taskChannel:
				if !ok {
					return
				}
				fmt.Printf("Worker number: %d task number: %s\n", w.id, task)
			case <-w.stop:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	close(w.stop)
}

func main() {
	pool := NewWorkerPool()
	pool.Start()

	pool.AddWorker()
	pool.AddWorker()
	time.Sleep(2 * time.Second)
	for i := 1; i <= 20; i++ {
		pool.AddTask(fmt.Sprintf("Task %d", i))
	}

	time.Sleep(1 * time.Second)
	pool.AddWorker()

	for i := 1; i <= 20; i++ {
		pool.AddTask(fmt.Sprintf("Task %d", i))
	}

	time.Sleep(1 * time.Second)
	pool.RemoveWorker()

	time.Sleep(1 * time.Second)
	pool.Stop()
}
