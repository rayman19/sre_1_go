package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/weather", func(w http.ResponseWriter, _ *http.Request) {
		if rand.Intn(2) == 0 {
			codes := []int{500, 502, 503, 504}
			w.WriteHeader(codes[rand.Intn(len(codes))])
			fmt.Fprintln(w, "Internal Server Error")
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"city": "Moscow", "value": 22}`)
		}
	})

	fmt.Println("Weather service running on :8080")
	http.ListenAndServe(":8080", nil)
}
