package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	defaultPort    = "8080"
	defaultTimeout = 10 * time.Second
	surflineAPI    = "https://services.surfline.com/search/site"
)

var (
	httpClient = &http.Client{Timeout: defaultTimeout}
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(loggingMiddleware(handleSearch)),
	}

	go func() {
		log.Println("Server started on", port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	// Graceful shutdowns with a timeout.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s %s Completed in %v", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	resp, err := httpClient.Get(surflineAPI + "?q=" + query + "&querySize=10&suggestionSize=10&newsSearch=true")
	if err != nil {
		log.Printf("Failed to fetch data from Surfline API: %v", err)
		http.Error(w, "Error fetching data from Surfline API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Surfline API returned unexpected status: %v", resp.Status)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response from Surfline API: %v", err)
		http.Error(w, "Error reading response from Surfline API", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
