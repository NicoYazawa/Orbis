package phase0

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var Version = "dev"

func Router(service string, checks map[string]Probe) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "orbis_http_requests_total", Help: "HTTP requests handled"}, []string{"service", "method", "route", "status"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "orbis_http_request_duration_seconds", Help: "HTTP request duration", Buckets: prometheus.DefBuckets}, []string{"service", "method", "route"})
	registry.MustRegister(requests, duration, prometheus.NewGoCollector())
	r.Use(func(c *gin.Context) {
		start := time.Now()
		ctx, span := otel.Tracer("orbis/http").Start(c.Request.Context(), c.Request.Method+" request")
		c.Request = c.Request.WithContext(ctx)
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			c.Header("X-Trace-ID", sc.TraceID().String())
		}
		c.Next()
		span.End()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		requests.WithLabelValues(service, c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
		duration.WithLabelValues(service, c.Request.Method, route).Observe(time.Since(start).Seconds())
		attrs := []any{"service", service, "method", c.Request.Method, "route", route, "status", c.Writer.Status(), "duration_ms", time.Since(start).Milliseconds()}
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			attrs = append(attrs, "trace_id", sc.TraceID().String(), "span_id", sc.SpanID().String())
		}
		slog.Info("http request", attrs...)
	})
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", gin.WrapH(ReadyHandler(checks)))
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	r.GET("/api/v1/system", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": service, "version": Version, "status": "ok"})
	})
	return r
}
