package phase0

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/crypto/bcrypt"
)

func TestConfigValidation(t *testing.T) {
	base := map[string]string{
		"ORBIS_DATABASE_URL": "postgres://u:p@localhost/db", "ORBIS_REDIS_ADDR": "localhost:6379",
		"ORBIS_S3_ENDPOINT": "localhost:9000", "ORBIS_S3_ACCESS_KEY": "key", "ORBIS_S3_SECRET_KEY": "secret",
		"ORBIS_S3_BUCKET": "orbis",
	}
	get := func(k string) string { return base[k] }
	if _, err := LoadConfig(get); err != nil {
		t.Fatal(err)
	}
	delete(base, "ORBIS_S3_SECRET_KEY")
	if _, err := LoadConfig(get); err == nil || !strings.Contains(err.Error(), "ORBIS_S3_SECRET_KEY") {
		t.Fatalf("expected missing secret error, got %v", err)
	}
	base["ORBIS_S3_SECRET_KEY"] = "secret"
	base["ORBIS_DATABASE_URL"] = "postgres://u:top-secret@%/db"
	if _, err := LoadConfig(get); err == nil || strings.Contains(err.Error(), "top-secret") {
		t.Fatalf("database URL leaked in error: %v", err)
	}
	base["ORBIS_DATABASE_URL"] = "postgres://u:p@localhost/db"
	base["ORBIS_REQUIRE_RAG"] = "true"
	if _, err := LoadConfig(get); err == nil || !strings.Contains(err.Error(), "ORBIS_QDRANT_URL") {
		t.Fatalf("expected required qdrant error, got %v", err)
	}
}

func TestReadinessTimesOutUnresponsiveProbe(t *testing.T) {
	checks := map[string]Probe{"postgres": fakeProbe{}, "stuck": ProbeFunc(func(context.Context) error { time.Sleep(5 * time.Second); return nil })}
	start := time.Now()
	rec := httptest.NewRecorder()
	ReadyHandler(checks).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable || time.Since(start) > 4*time.Second || !strings.Contains(rec.Body.String(), `"stuck":"timeout"`) {
		t.Fatalf("response=%d %s duration=%s", rec.Code, rec.Body.String(), time.Since(start))
	}
}

func TestRouterEndpoints(t *testing.T) {
	r := Router("api", map[string]Probe{"postgres": fakeProbe{}})
	for path, expected := range map[string]int{"/healthz": 200, "/readyz": 200, "/metrics": 200, "/api/v1/system": 200, "/missing": 404} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != expected {
			t.Fatalf("%s status=%d", path, rec.Code)
		}
		if path == "/api/v1/system" && !strings.Contains(rec.Body.String(), `"service":"api"`) {
			t.Fatalf("system response %s", rec.Body.String())
		}
	}
}

func TestQdrantURLPreservesQuery(t *testing.T) {
	got, err := qdrantURL("http://qdrant:6333", "/collections/check/points?wait=true")
	if err != nil || got != "http://qdrant:6333/collections/check/points?wait=true" {
		t.Fatalf("got %s, %v", got, err)
	}
}

func TestMissingObjectOnlyAcceptsNotFound(t *testing.T) {
	if !isMissingObject(minio.ErrorResponse{StatusCode: 404, Code: "NoSuchKey"}) {
		t.Fatal("404 should be missing")
	}
	if isMissingObject(errors.New("connection refused")) || isMissingObject(minio.ErrorResponse{StatusCode: 403, Code: "AccessDenied"}) || isMissingObject(nil) {
		t.Fatal("non-404 accepted as deletion")
	}
}

func TestOTLPOriginAppendsTracePathAndService(t *testing.T) {
	type request struct {
		path    string
		payload []byte
	}
	requests := make(chan request, 1)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := io.ReadAll(r.Body)
		requests <- request{r.URL.Path, payload}
		w.WriteHeader(http.StatusOK)
	}))
	defer receiver.Close()
	shutdown, err := SetupTelemetry(context.Background(), "api", receiver.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, span := otel.Tracer("test").Start(context.Background(), "probe")
	span.End()
	tp := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	if err := tp.ForceFlush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	got := <-requests
	if got.path != "/v1/traces" || !bytes.Contains(got.payload, []byte("orbis-api")) {
		t.Fatalf("OTLP path=%s service absent=%v", got.path, !bytes.Contains(got.payload, []byte("orbis-api")))
	}
}

func TestBrowserPresignUsesInternalBucketLocation(t *testing.T) {
	var locations atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["location"]; ok {
			locations.Add(1)
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">us-east-1</LocationConstraint>`))
			return
		}
		if r.Method == http.MethodPut {
			w.Header().Set("ETag", `"test"`)
			w.WriteHeader(http.StatusOK)
			return
		}
		t.Errorf("unexpected internal S3 request %s %s", r.Method, r.URL)
	}))
	defer server.Close()
	cfg := Config{S3Endpoint: server.URL, S3PublicEndpoint: "http://127.0.0.1:1", S3AccessKey: "key", S3SecretKey: "secret", S3Bucket: "orbis-phase0"}
	internal, err := newS3Client(cfg.S3Endpoint, cfg)
	if err != nil {
		t.Fatal(err)
	}
	public, err := newS3Client(cfg.S3PublicEndpoint, cfg)
	if err != nil {
		t.Fatal(err)
	}
	d := &Dependencies{S3: internal, PublicS3: public, Config: cfg}
	result, err := d.CreateBrowserPresign(context.Background(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if locations.Load() == 0 || !strings.HasPrefix(result.PresignedURL, "http://127.0.0.1:1/") {
		t.Fatalf("locations=%d URL=%s", locations.Load(), result.PresignedURL)
	}
}

type fakeProbe struct{ err error }

func (p fakeProbe) Check(context.Context) error { return p.err }

func TestReadinessReportsDependencyFailure(t *testing.T) {
	checks := map[string]Probe{"postgres": fakeProbe{}, "redis": fakeProbe{err: errors.New("down")}, "s3": fakeProbe{}}
	h := ReadyHandler(checks)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"redis":"down"`) {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
	checks["redis"] = fakeProbe{}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
}

func TestEnsureAdminDoesNotResetExistingPassword(t *testing.T) {
	store := &fakeAdminStore{}
	if err := EnsureAdmin(context.Background(), store, "admin", "first-password"); err != nil {
		t.Fatal(err)
	}
	first := store.hash
	if first == "first-password" || first == "" {
		t.Fatalf("password not hashed: %q", first)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(first), []byte("first-password")); err != nil {
		t.Fatalf("invalid bcrypt hash: %v", err)
	}
	if err := EnsureAdmin(context.Background(), store, "admin", "second-password"); err != nil {
		t.Fatal(err)
	}
	if store.creates != 1 || store.hash != first {
		t.Fatalf("rerun modified admin: %+v", store)
	}
	if err := EnsureAdmin(context.Background(), store, "admin", ""); err == nil {
		t.Fatal("empty password accepted")
	}
}

type fakeAdminStore struct {
	hash    string
	creates int
}

func (s *fakeAdminStore) CreateAdminIfAbsent(_ context.Context, _ string, hash string) error {
	if s.creates == 0 {
		s.hash = hash
		s.creates++
	}
	return nil
}
