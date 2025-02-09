package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := GetData("http://localhost:8080/weather")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Weather data: "+data)
	})

	fmt.Println("Client service running on :8081")
	http.ListenAndServe(":8081", nil)
}

func GetData(url string) (string, error) {
	const maxAttempts = 3
	client := &http.Client{Timeout: 5 * time.Second}

	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		resp, err := client.Get(url)
		if err != nil {
			return "", fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("failed to read body: %w", err)
			}
			return string(data), nil
		}

		if isRetryable(resp.StatusCode) {
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			if attempt < maxAttempts-1 {
				switch attempt {
				case 1:
					time.Sleep(1 * time.Second)
				case 2:
					time.Sleep(2 * time.Second)
				}
				continue
			}
		} else {
			return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
	}

	return "", fmt.Errorf("after %d attempts: %w", maxAttempts, lastErr)
}

func isRetryable(code int) bool {
	switch code {
	case 500, 502, 503, 504:
		return true
	default:
		return false
	}
}
