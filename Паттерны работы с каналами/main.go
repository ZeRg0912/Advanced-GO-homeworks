package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

const workersCount = 5

type Job struct {
	ID  int
	URL string
}

type Result struct {
	Job      Job
	Status   string
	Duration time.Duration
	WorkerID int
}

func worker(workerID int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		result := processURL(workerID, job)
		results <- result
	}
}

func processURL(workerID int, job Job) Result {
	start := time.Now()

	randomDelay := time.Duration(rand.Intn(900)+100) * time.Millisecond
	time.Sleep(randomDelay)

	duration := time.Since(start)

	return Result{
		Job:      job,
		Status:   "обработан",
		Duration: duration,
		WorkerID: workerID,
	}
}

func printReport(results []Result) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Job.ID < results[j].Job.ID
	})

	var totalDuration time.Duration
	successCount := 0

	fmt.Println("Итоговый отчёт")
	fmt.Println("==============")

	for _, result := range results {
		fmt.Printf(
			"ID: %02d | URL: %-35s | Статус: %-10s | Время: %-10s | Worker: %d\n",
			result.Job.ID,
			result.Job.URL,
			result.Status,
			result.Duration.Round(time.Millisecond),
			result.WorkerID,
		)

		totalDuration += result.Duration

		if result.Status == "обработан" {
			successCount++
		}
	}

	fmt.Println("==============")
	fmt.Printf("Всего URL: %d\n", len(results))
	fmt.Printf("Успешно обработано: %d\n", successCount)

	if len(results) > 0 {
		averageDuration := totalDuration / time.Duration(len(results))
		fmt.Printf("Среднее время обработки: %s\n", averageDuration.Round(time.Millisecond))
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	urls := []string{
		"https://example.com",
		"https://google.com",
		"https://github.com",
		"https://golang.org",
		"https://stackoverflow.com",
		"https://habr.com",
		"https://gitlab.com",
		"https://pkg.go.dev",
		"https://go.dev",
		"https://wikipedia.org",
		"https://reddit.com",
		"https://news.ycombinator.com",
	}

	jobs := make(chan Job, len(urls))
	results := make(chan Result, len(urls))

	var wg sync.WaitGroup

	for i := 1; i <= workersCount; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	for i, url := range urls {
		jobs <- Job{
			ID:  i + 1,
			URL: url,
		}
	}

	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	collectedResults := make([]Result, 0, len(urls))

	for result := range results {
		collectedResults = append(collectedResults, result)
	}

	printReport(collectedResults)
}
