package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

const jobsCount, resultCount = 100, 100

var actions = []string{"logged in", "logged out", "created record", "deleted record", "updated account"}

type logItem struct {
	action    string
	timestamp time.Time
}

type User struct {
	id    int
	email string
	logs  []logItem
}

func (u User) getActivityInfo() string {
	output := fmt.Sprintf("UID: %d; Email: %s;\nActivity Log:\n", u.id, u.email)
	for index, item := range u.logs {
		output += fmt.Sprintf("%d. [%s] at %s\n", index, item.action, item.timestamp.Format(time.RFC3339))
	}

	return output
}

func main() {
	var wg sync.WaitGroup
	jobs := make(chan int)
	results := make(chan User)

	rand.Seed(time.Now().Unix())

	startTime := time.Now()

	generateUsers(100, jobs, results)

	worker(jobsCount, jobs, &wg)

	saveUserInfo(resultCount, results, &wg)

	wg.Wait()

	fmt.Printf("DONE! Time Elapsed: %.2f seconds\n", time.Since(startTime).Seconds())
}

func worker(jobsCount int, jobs chan<- int, wg *sync.WaitGroup) {

	for i := 0; i < jobsCount; i++ {
		wg.Add(1)
		jobs <- i
	}
}

func saveUserInfo(resultCount int, result chan User, wg *sync.WaitGroup) {
	for i := 0; i < resultCount; i++ {
		go func() {
			for user := range result {
				fmt.Printf("WRITING FILE FOR UID %d\n", user.id)
				err := os.MkdirAll("users", os.ModePerm)
				if err != nil {
					log.Fatalf("Ошибка при создании папки users: %v", err)
				}

				filename := fmt.Sprintf("users/uid%d.txt", user.id)
				file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					log.Fatal(err)
				}

				defer file.Close() // Гарантируем закрытие файла после записи

				_, err = file.WriteString(user.getActivityInfo())
				if err != nil {
					log.Fatalf("Ошибка при записи в файл: %v", err)

				}
				time.Sleep(time.Second)
				wg.Done()
			}
			close(result)
		}()
	}

}

func generateUsers(count int, jobs <-chan int, results chan User) {

	for i := 0; i < count; i++ {
		go func() {
			for j := range jobs {
				results <- User{
					id:    j,
					email: fmt.Sprintf("user%d@company.com", j),
					logs:  generateLogs(rand.Intn(1000)),
				}
				fmt.Printf("generated user %d\n", j)
				time.Sleep(time.Millisecond * 100)
			}
			close(results)
		}()
	}
}

func generateLogs(count int) []logItem {
	logs := make([]logItem, count)

	for i := 0; i < count; i++ {
		logs[i] = logItem{
			action:    actions[rand.Intn(len(actions)-1)],
			timestamp: time.Now(),
		}
	}

	return logs
}
