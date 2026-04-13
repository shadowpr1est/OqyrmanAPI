package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var registerMetricsOnce sync.Once

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "oqyrman",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)
	requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "oqyrman",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)
	inFlightRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "oqyrman",
			Subsystem: "http",
			Name:      "in_flight_requests",
			Help:      "Number of HTTP requests currently in flight.",
		},
		[]string{"method", "route"},
	)
)

func registerMetrics() {
	registerMetricsOnce.Do(func() {
		prometheus.MustRegister(requestDuration)
		prometheus.MustRegister(requestTotal)
		prometheus.MustRegister(inFlightRequests)
	})
}

// Metrics exposes Prometheus request metrics for the HTTP API.
func Metrics() gin.HandlerFunc {
	registerMetrics()

	return func(c *gin.Context) {
		start := time.Now()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		inFlightRequests.WithLabelValues(c.Request.Method, route).Inc()
		defer inFlightRequests.WithLabelValues(c.Request.Method, route).Dec()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		labels := []string{c.Request.Method, route, status}
		requestDuration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
		requestTotal.WithLabelValues(labels...).Inc()
	}
}