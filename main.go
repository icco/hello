package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/icco/gutil/logging"
)

var (
	service = "hello"
	project = "icco-cloud"
	log     = logging.Must(logging.NewLogger(service))
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Infow("Starting up", "host", fmt.Sprintf("http://localhost:%s", port))

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(logging.Middleware(log.Desugar(), project))

	r.Get("/", hello)
	r.Get("/healthz", hello)
	r.HandleFunc("/204", twoOhFour)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

type helloRespJSON struct {
	Status  string `json:"status"`
	Message string `json:"msg"`
}

func hello(w http.ResponseWriter, r *http.Request) {
	resp := helloRespJSON{"ok", "Hello World"}

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
