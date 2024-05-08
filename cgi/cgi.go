package main

import (
	"fmt"
	"net/http"
	"net/http/cgi"
	"time"
)

func processPart(part int, results chan<- string) {
	time.Sleep(5 * time.Second)
	results <- fmt.Sprintf("Processed part: %d", part)
}

func main() {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		results := make(chan string, 5) // Buffered channel for results

		// Start several goroutines
		for i := 1; i <= 5; i++ {
			go processPart(i, results)
		}

		// Gather results
		for j := 1; j <= 5; j++ {
			fmt.Fprintf(w, "%s\n", <-results)
		}
	}))
}
