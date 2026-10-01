package phase0

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Probe interface{ Check(context.Context) error }
type ProbeFunc func(context.Context) error

func (f ProbeFunc) Check(ctx context.Context) error { return f(ctx) }

func ReadyHandler(checks map[string]Probe) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		failures := map[string]string{}
		seen := map[string]bool{}
		results := make(chan struct {
			name string
			err  error
		}, len(checks))
		for name, probe := range checks {
			name, probe := name, probe
			go func() {
				results <- struct {
					name string
					err  error
				}{name, probe.Check(ctx)}
			}()
		}
		for range checks {
			select {
			case result := <-results:
				seen[result.name] = true
				if result.err != nil {
					failures[result.name] = "down"
				}
			case <-ctx.Done():
				for name := range checks {
					if !seen[name] {
						failures[name] = "timeout"
					}
				}
				goto done
			}
		}
	done:
		w.Header().Set("Content-Type", "application/json")
		if len(failures) > 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "degraded", "dependencies": failures})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ready"})
	})
}
