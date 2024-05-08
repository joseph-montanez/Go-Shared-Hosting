package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/fcgi"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	ticker *time.Ticker
	quit   chan struct{}
	mutex  sync.Mutex
)

func processPart(part int, results chan<- string) {
	time.Sleep(5 * time.Second)
	results <- fmt.Sprintf("Processed part: %d", part)
}

func main() {
	r := mux.NewRouter()

	http.Handle("/", dynamicPathAdjustMiddleware(r))

	r.HandleFunc("/", safeHandler(HomeHandler))
	r.HandleFunc("/hello", safeHandler(TemplHandler))
	r.HandleFunc("/async", safeHandler(AsyncHandler))
	r.HandleFunc("/cookie", safeHandler(CookieHandler))
	r.HandleFunc("/start", safeHandler(StartLoggingHandler))
	r.HandleFunc("/stop", safeHandler(StopLoggingHandler))
	r.HandleFunc("/events", safeHandler(sseHandler))

	quit = make(chan struct{})

	if err := fcgi.Serve(nil, http.DefaultServeMux); err != nil {
		fmt.Println("Error serving FastCGI:", err)
	}
}

func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Recovered from panic: %v", err)
				http.Error(w, "Internal Server Error", 500)
			}
		}()
		fn(w, r)
	}
}

func dynamicPathAdjustMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Print incoming path for debugging
		//fmt.Printf("Original URL Path: %s\n", r.URL.Path)

		// Automatically find and trim up to '/main.fcgi'
		splitPath := strings.SplitN(r.URL.Path, "/main.fcgi", 2)
		if len(splitPath) > 1 {
			newPath := splitPath[1]
			if newPath == "" || newPath[0] != '/' {
				newPath = "/" + newPath
			}
			r.URL.Path = newPath
			//fmt.Printf("Adjusted URL Path to: %s\n", r.URL.Path)
		}

		// Proceed with the modified request
		next.ServeHTTP(w, r)
	})
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	pid := os.Getpid()

	fmt.Fprintln(w, "Welcome to the Home Page! - "+strconv.Itoa(pid))
}

func CookieHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:  "testcookie",
		Value: "123456",
		Path:  "/",
	})
	fmt.Fprintln(w, "Cookie set!")
}

func AsyncHandler(w http.ResponseWriter, r *http.Request) {
	results := make(chan string, 5)

	for i := 1; i <= 5; i++ {
		go processPart(i, results)
	}

	for j := 1; j <= 5; j++ {
		fmt.Fprintf(w, "%s\n", <-results)
	}
}

func StartLoggingHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	if ticker == nil {
		ticker = time.NewTicker(10 * time.Second)
		go func() {
			for {
				select {
				case t := <-ticker.C:
					writeTimeToFile(t)
				case <-quit:
					ticker.Stop()
					ticker = nil
					return
				}
			}
		}()
		fmt.Fprintln(w, "Started logging to file.")
	} else {
		fmt.Fprintln(w, "Logging already running.")
	}
}

func StopLoggingHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	if ticker != nil {
		close(quit)
		quit = make(chan struct{})
		fmt.Fprintln(w, "Stopped logging to file.")
	} else {
		fmt.Fprintln(w, "No logging is active.")
	}
}

func writeTimeToFile(t time.Time) {
	file, err := os.OpenFile("time_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	if _, err = file.WriteString(t.Format("2006-01-02 15:04:05") + "\n"); err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func TemplHandler(w http.ResponseWriter, r *http.Request) {
	pid := os.Getpid()

	component := hello(strconv.Itoa(pid))
	component.Render(context.Background(), w)
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/events - sseHandler\n")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher.Flush()

	//time.Sleep(5 * time.Second)
	//fmt.Fprintf(w, "data: 5.00\n\n")
	//flusher.Flush()
	//
	//time.Sleep(5 * time.Second)
	//fmt.Fprintf(w, "data: 10.00\n\n")
	//flusher.Flush()

	for {
		select {
		case t := <-ticker.C:
			// Debug output
			fmt.Printf("Sending time: %s\n", t.String())
			// Write to the ResponseWriter
			fmt.Fprintf(w, ":heartbeat\n\n")
			fmt.Fprintf(w, "data: %s\n\n", t.Format(time.RFC1123))
			// Flush the data immediately instead of buffering it
			flusher.Flush()
		case <-r.Context().Done():
			// Debug output
			fmt.Println("Client closed connection")
			return // Exit if client closes connection
		}
	}
}
