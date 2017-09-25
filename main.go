package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	// "gopkg.in/unrolled/secure.v1"
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	server := http.NewServeMux()
	server.HandleFunc("/", hello)
	server.HandleFunc("/204", twoOhFour)

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}

type HelloRespJson struct {
	Status  string `json:"status"`
	Message string `json:"msg"`
}

func hello(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request: %s", r.URL.Path)
	resp := HelloRespJson{"ok", "Hello World"}

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func twoOhFour(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
