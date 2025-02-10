package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	state           State
	failureCount    int
	resetTimeout    time.Duration
	lastFailureTime time.Time
	mutex           sync.Mutex
}

func NewCircuitBreaker(resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        Closed,
		resetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()

	switch cb.state {
	case Closed:
		return true
	case Open:
		if now.Sub(cb.lastFailureTime) >= cb.resetTimeout {
			cb.state = HalfOpen
			log.Println("Circuit Breaker state: Open → Half-Open")
			return true // Разрешаем тестовый запрос
		}
		return false
	case HalfOpen:
		return false // Разрешаем только один запрос, который уже в процессе
	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case HalfOpen:
		cb.state = Closed
		cb.failureCount = 0
		log.Println("Circuit Breaker state: Half-Open → Closed")
	case Closed:
		cb.failureCount = 0 // Сброс счетчика при успехе
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case Closed:
		if cb.failureCount >= 3 {
			cb.state = Open
			log.Println("Circuit Breaker state: Closed → Open")
		}
	case HalfOpen:
		cb.state = Open
		log.Println("Circuit Breaker state: Half-Open → Open")
	}
}

var cb = NewCircuitBreaker(10 * time.Second)

func GetDataWithCircuitBreaker(url string) (string, error) {
	if !cb.AllowRequest() {
		return "", fmt.Errorf("circuit breaker blocked the request")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		cb.RecordFailure()
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if isRetryable(resp.StatusCode) {
			cb.RecordFailure()
		}
		return "", fmt.Errorf("API error: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	cb.RecordSuccess()
	return string(data), nil
}

func isRetryable(code int) bool {
	switch code {
	case 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		data, err := GetDataWithCircuitBreaker("http://localhost:8080/weather")
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Response: %s", data)
	})

	log.Println("Client with Circuit Breaker running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
