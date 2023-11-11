package main

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

const numPasswords = 100000

func hashSequential(passwords []string) []string {
	var hashedPasswords []string
	for _, password := range passwords {
		hashed := sha256.Sum256([]byte(password))
		hashedPasswords = append(hashedPasswords, fmt.Sprintf("%x", hashed))
	}
	return hashedPasswords
}

const numWorkers = 4

var workerPool sync.Pool

func init() {
	workerPool.New = func() interface{} {
		return make(chan string, 1)
	}
}

func worker(wg *sync.WaitGroup, passwordsChannel <-chan string, result chan<- string) {
	defer wg.Done()
	for password := range passwordsChannel {
		hashed := sha256.Sum256([]byte(password))
		result <- fmt.Sprintf("%x", hashed)
	}
}

func hashParallel(passwords []string) []string {
	var (
		wg              sync.WaitGroup
		mutex           sync.Mutex
		hashedPasswords []string
	)

	passwordsChannel := make(chan string, numWorkers)
	resultChannel := make(chan string, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, passwordsChannel, resultChannel)
	}

	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	go func() {
		for _, password := range passwords {
			passwordsChannel <- password
		}
		close(passwordsChannel)
	}()

	for hashed := range resultChannel {
		mutex.Lock()
		hashedPasswords = append(hashedPasswords, hashed)
		mutex.Unlock()
	}

	return hashedPasswords
}

func main() {
	var passwords []string
	for i := 0; i < numPasswords; i++ {
		passwords = append(passwords, fmt.Sprintf("password%d", i))
	}

	startTimeSequential := time.Now()
	hashSequential(passwords)
	elapsedTimeSequential := time.Since(startTimeSequential)
	fmt.Printf("Sequential: Elapsed time: %s\n", elapsedTimeSequential)

	startTimeParallel := time.Now()
	hashParallel(passwords)
	elapsedTimeParallel := time.Since(startTimeParallel)
	fmt.Printf("Parallel: Elapsed time: %s\n", elapsedTimeParallel)

}
