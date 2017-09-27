package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/justinas/alice"
	"gopkg.in/unrolled/secure.v1"
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	development := os.Getenv("RACK_ENV") == "development"
	s := secure.New(secure.Options{
		AllowedHosts:          []string{"hello.natwelch.com"},
		HostsProxyHeaders:     []string{"X-Forwarded-Host"},
		SSLRedirect:           true,
		SSLHost:               "hello.natwelch.com",
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		STSIncludeSubdomains:  false,
		STSPreload:            false,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "same-origin",
		IsDevelopment:         development,
	})
	secHandler := s.Handler

	commonHandlers := alice.New(secHandler)

	server := http.NewServeMux()
	server.Handle("/", commonHandlers.ThenFunc(hello))
	server.Handle("/healthz", hello)
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
