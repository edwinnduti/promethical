package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"path"},
)

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		totalRequests.WithLabelValues(path).Inc()
	})
}

func init() {
	prometheus.Register(totalRequests)
}

func main() {
	router := mux.NewRouter()
	router.Use(prometheusMiddleware)

	// Prometheus endpoint
	router.Path("/prometheus").Handler(promhttp.Handler())

	// Serving static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	fmt.Println("Serving requests on port 9000")
	err := http.ListenAndServe(":9000", router)
	log.Fatal(err)
}