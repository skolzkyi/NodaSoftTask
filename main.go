package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A Ttype represents a meaninglessness of our life
type Ttype struct {
	id         string
	cT         string // время создания
	fT         string // время выполнения
	taskRESULT []byte
}

func main() {
	workers := 10 //num of workers
	signal := make(chan struct{})
	appWorkTime := time.Second * 3 //time of work application

	superChan := make(chan Ttype, 10)

	doneTasks := make(chan Ttype) //left two channels, perhaps this architecture was meant
	undoneTasks := make(chan error)

	result := map[string]Ttype{}
	err := []error{}

	go func() {
		t := time.NewTimer(appWorkTime)
		<-t.C
		close(signal)
	}()

	go func() {
		defer close(superChan)
		for {
			select {
			case <-signal:
				return
			default:
				ttime := time.Now() //only one syscall
				ft := ttime.Format(time.RFC3339)
				ns := ttime.Nanosecond()
				if ns%2 > 0 { // вот такое условие появления ошибочных тасков
					ft = ft + " - Some error occured"
				}
				idlabel := strconv.Itoa(int(ttime.Unix())) + "_" + strconv.Itoa(ns)
				superChan <- Ttype{cT: ft, id: idlabel} // передаем таск на выполнение
			}
		}
	}()

	wgw := sync.WaitGroup{}
	wgw.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wgw.Done()
			for data := range superChan {
				data.fT = time.Now().Format(time.RFC3339Nano)
				_, err := time.Parse(time.RFC3339, data.cT)
				if err != nil {
					data.taskRESULT = []byte("task has been successed")
					doneTasks <- data
				} else {
					data.taskRESULT = []byte("something went wrong")
					undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", data.id, data.cT, data.taskRESULT)
				}

				time.Sleep(time.Millisecond * 150)
			}
		}()

	}
	go func() {
		wgw.Wait()
		close(doneTasks)
		close(undoneTasks)
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for r := range doneTasks {
			result[r.id] = r
		}
	}()
	go func() {
		defer wg.Done()
		for r := range undoneTasks {
			err = append(err, r)
		}
	}()

	wg.Wait()

	println("Errors:")
	for _, r := range err {
		println(r.Error())
	}

	println("Done tasks:")
	for r := range result {
		println(r)
	}
}
