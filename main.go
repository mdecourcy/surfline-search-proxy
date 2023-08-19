package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

const (
	surflineAPI = "https://services.surfline.com/search/site"
	port        = ":8080"
)

func main() {
	http.HandleFunc("/search", handleSearch)

	log.Println("Server started on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(surflineAPI + "?q=" + query + "&querySize=10&suggestionSize=10&newsSearch=true")
	if err != nil {
		http.Error(w, "Error fetching data from Surfline API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response from Surfline API", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
