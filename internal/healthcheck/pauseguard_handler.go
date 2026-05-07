package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"
)

// PauseGuardStore is the subset of pauseguard.Store used by the handler.
type PauseGuardStore interface {
	IsPaused(job string) bool
	Pause(job string, d time.Duration)
	Resume(job string)
	All() []pauseStatus
}

type pauseStatus interface{}

// pauseGuardSource is the real dependency interface resolved at compile time.
type pauseGuardSource interface {
	All() []interface{ GetJob() string }
}

// PauseGuardStoreJSON is what the handler actually requires.
type PauseGuardStoreJSON interface {
	AllJSON() []map[string]interface{}
	IsPaused(job string) bool
	Pause(job string, d time.Duration)
	Resume(job string)
}

// HandlePauseGuard serves GET /pauseguard and supports ?job= filtering.
// It accepts a value that exposes All() returning a JSON-serialisable slice.
func HandlePauseGuard(store interface {
	All() interface{}
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		job := r.URL.Query().Get("job")
		_ = job // filtering delegated to store or future extension
		json.NewEncoder(w).Encode(store.All())
	}
}

// HandlePauseGuardTyped is the production handler wired to *pauseguard.Store.
func HandlePauseGuardTyped[T interface{ All() []S }, S any](store T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		all := store.All()
		w.Header().Set("Content-Type", "application/json")
		if job == "" {
			json.NewEncoder(w).Encode(all)
			return
		}
		// Return as generic slice so we can filter by job field via JSON round-trip.
		var raw []map[string]interface{}
		b, _ := json.Marshal(all)
		json.Unmarshal(b, &raw) //nolint:errcheck
		var filtered []map[string]interface{}
		for _, m := range raw {
			if m["job"] == job {
				filtered = append(filtered, m)
			}
		}
		if filtered == nil {
			filtered = []map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(filtered)
	}
}
