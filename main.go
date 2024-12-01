package main

import (
	"fmt"
	"net/http"
	"os"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/icco/gutil/logging"
	"github.com/unrolled/render"
	"github.com/unrolled/secure"
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

	secureMiddleware := secure.New(secure.Options{
		SSLRedirect:           false,
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com; connect-src 'self' https://reportd.natwelch.com; report-uri https://reportd.natwelch.com/report/hello; report-to default",
	})

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(logging.Middleware(log.Desugar(), project))
	r.Use(secureMiddleware.Handler)

	r.Get("/", hello("html"))
	r.Get("/json", hello("json"))
	r.Get("/healthz", hello("json"))
	r.HandleFunc("/204", twoOhFour)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

type helloRespJSON struct {
	Status  string `json:"status"`
	Message string `json:"msg"`
}

func hello(format string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := helloRespJSON{"ok", "Hello World"}
		re := render.New()

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("report-to", `{"group":"default","max_age":10886400,"endpoints":[{"url":"https://reportd.natwelch.com/report/hello"}]}`)
		w.Header().Set("reporting-endpoints", `default="https://reportd.natwelch.com/reporting/hello"`)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("nel", `{"report_to":"default","max_age":2592000}`)
		w.Header().Set("Feature-Policy", "geolocation 'none'; midi 'none'; sync-xhr 'none'; microphone 'none'; camera 'none'; magnetometer 'none'; gyroscope 'none'; speaker 'none'; fullscreen 'self'; payment 'none'; usb 'none'")
		w.Header().Set("Referrer-Policy", "no-referrer")

		switch format {
		case "json":
			re.JSON(w, http.StatusOK, resp)
		case "html":
			re.HTML(w, http.StatusOK, "hello", resp)
		default:
			http.Error(w, "invalid format", http.StatusBadRequest)
		}
	}
}

func twoOhFour(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
