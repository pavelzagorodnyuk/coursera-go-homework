package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func main() {
	// functions are triggered by tests
}

// ExecutePipeline runs job functions as a pipeline
func ExecutePipeline(jobs ...job) {

	if len(jobs) < 2 {
		return
	}

	channels := make([]chan interface{}, len(jobs)-1)
	for i := range channels {
		channels[i] = make(chan interface{})
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(jobs))

	go func() {
		jobs[0](nil, channels[0])
		wg.Done()
		close(channels[0])
	}()

	for i := 1; i < len(jobs)-1; i++ {
		go func(i int) {
			jobs[i](channels[i-1], channels[i])
			wg.Done()
			close(channels[i])
		}(i)
	}

	go func() {
		jobs[len(jobs)-1](channels[len(channels)-1], nil)
		wg.Done()
	}()

	wg.Wait()
}

var (
	quota       = make(chan struct{}, 1)
	workerCount = 7
)

func SingleHash(in, out chan interface{}) {

	singleHashWG := &sync.WaitGroup{}
	singleHashWG.Add(workerCount)

	// starting a worker pool
	for i := 0; i < workerCount; i++ {
		go func(in, out chan interface{}, singleHashWG *sync.WaitGroup) {
			defer singleHashWG.Done()

			// processing incoming data
			for source := range in {
				intValue, ok := source.(int)
				if !ok {
					panic("can`t convert source to int")
				}

				data := strconv.Itoa(intValue)

				parts := make([]string, 2)
				channel := make(chan string, 1)

				go func() {
					channel <- DataSignerCrc32(data)
				}()

				quota <- struct{}{}
				md5 := DataSignerMd5(data)
				<-quota

				parts[1] = DataSignerCrc32(md5)
				parts[0] = <-channel

				out <- strings.Join(parts, "~")
			}
		}(in, out, singleHashWG)
	}

	singleHashWG.Wait()
}

const multiHashTh = 6

func MultiHash(in, out chan interface{}) {

	multiHashWG := &sync.WaitGroup{}
	multiHashWG.Add(workerCount)

	// starting a worker pool
	for i := 0; i < workerCount; i++ {
		go func(in, out chan interface{}, multiHashWG *sync.WaitGroup) {
			defer multiHashWG.Done()

			// processing incoming data
			for source := range in {
				data, ok := source.(string)
				if !ok {
					panic("can`t convert source to string")
				}

				wg := &sync.WaitGroup{}
				wg.Add(multiHashTh)

				var parts [multiHashTh]string

				for i := range parts {
					go func(index int) {
						defer wg.Done()
						parts[index] = DataSignerCrc32(strconv.Itoa(index) + data)
					}(i)
				}

				wg.Wait()

				out <- strings.Join(parts[:], "")
			}
		}(in, out, multiHashWG)
	}

	multiHashWG.Wait()
}

func CombineResults(in, out chan interface{}) {

	var hashes []string

	for source := range in {
		value, ok := source.(string)
		if !ok {
			panic("can`t convert source to string")
		}

		hashes = append(hashes, value)
	}

	sort.StringSlice(hashes).Sort()

	out <- strings.Join(hashes, "_")
}
