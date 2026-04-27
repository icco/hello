// Command hello serves a tiny landing page and JSON greeting.
package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/icco/gutil/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/unrolled/render"
	"github.com/unrolled/secure"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"
)

//go:embed templates
var embeddedTemplates embed.FS

const service = "hello"

func main() {
	log := logging.Must(logging.NewLogger(service))
	defer func() {
		if err := log.Sync(); err != nil {
			log.Debugw("logger sync", zap.Error(err))
		}
	}()

	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	registry := prometheus.NewRegistry()
	exporter, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		log.Errorw("otel prometheus exporter", zap.Error(err))
		os.Exit(1)
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(mp)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mp.Shutdown(shutdownCtx); err != nil {
			log.Warnw("meter provider shutdown", zap.Error(err))
		}
	}()

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router(log, promhttp.HandlerFor(registry, promhttp.HandlerOpts{})),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Infow("http server starting", "addr", fmt.Sprintf("http://localhost:%s", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorw("http server", zap.Error(err))
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorw("http shutdown", zap.Error(err))
	}
}

// router builds the HTTP handler, wrapped with otelhttp (excluding /metrics).
func router(log *zap.SugaredLogger, metricsHandler http.Handler) http.Handler {
	secureMiddleware := secure.New(secure.Options{
		SSLRedirect:           false,
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com; connect-src 'self' https://reportd.natwelch.com; report-uri https://reportd.natwelch.com/report/hello; report-to default",
		ReferrerPolicy:        "no-referrer",
		FeaturePolicy:         "geolocation 'none'; midi 'none'; sync-xhr 'none'; microphone 'none'; camera 'none'; magnetometer 'none'; gyroscope 'none'; fullscreen 'none'; payment 'none'; usb 'none'",
	})

	r := chi.NewRouter()
	r.Use(logging.Middleware(log.Desugar()))
	r.Use(routeTag)
	r.Use(secureMiddleware.Handler)

	r.Get("/", hello("html"))
	r.Get("/json", hello("json"))
	r.Get("/healthz", hello("json"))
	r.HandleFunc("/204", twoOhFour)

	if metricsHandler != nil {
		r.Method(http.MethodGet, "/metrics", metricsHandler)
	}

	return otelhttp.NewHandler(r, service,
		otelhttp.WithFilter(func(req *http.Request) bool {
			return req.URL.Path != "/metrics"
		}),
	)
}

// routeTag stamps the chi route pattern onto otelhttp metric labels.
func routeTag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		labeler, ok := otelhttp.LabelerFromContext(r.Context())
		if !ok {
			return
		}
		if pattern := chi.RouteContext(r.Context()).RoutePattern(); pattern != "" {
			labeler.Add(semconv.HTTPRoute(pattern))
		}
	})
}

type helloRespJSON struct {
	Status  string `json:"status"`
	Message string `json:"msg"`
}

func hello(format string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logging.FromContext(r.Context())
		resp := helloRespJSON{"ok", "Hello World"}
		re := render.New(render.Options{
			FileSystem: &render.EmbedFileSystem{FS: embeddedTemplates},
		})

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("report-to", `{"group":"default","max_age":10886400,"endpoints":[{"url":"https://reportd.natwelch.com/report/hello"}]}`)
		w.Header().Set("reporting-endpoints", `default="https://reportd.natwelch.com/reporting/hello"`)
		w.Header().Set("nel", `{"report_to":"default","max_age":2592000}`)

		switch format {
		case "json":
			if err := re.JSON(w, http.StatusOK, resp); err != nil {
				l.Errorw("render json", zap.Error(err))
			}
		case "html":
			if err := re.HTML(w, http.StatusOK, "hello", resp); err != nil {
				l.Errorw("render html", zap.Error(err))
			}
		default:
			http.Error(w, "invalid format", http.StatusBadRequest)
		}
	}
}

func twoOhFour(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
