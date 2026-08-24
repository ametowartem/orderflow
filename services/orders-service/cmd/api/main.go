package main

import (
	"log"
	"net/http"
)

func health(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		log.Printf("health: failed to write response: %v", err)
	}
}

func main() {
	http.HandleFunc("/health", health)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
