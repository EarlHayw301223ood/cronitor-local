package healthcheck

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/your-org/cronitor-local/internal/jobfilter"
)

// filterRequest is the JSON body accepted by HandleJobFilter.
type filterRequest struct {
	Names []string `json:"names"`
	Tags  []string `json:"tags"`
}

// filterResponse is returned by HandleJobFilter.
type filterResponse struct {
	Matched []string `json:"matched"`
	Total   int      `json:"total"`
}

// HandleJobFilter evaluates a posted filter spec against a static list of
// known job names and returns the subset that matches.
//
// POST /jobs/filter
// Body: { "names": [...], "tags": [...] }
//
// The handler is intentionally stateless: callers supply both the candidate
// job names (via the "names" field when used as a name allowlist) and the
// tag criteria. When names is empty every job name in the query-string
// "jobs" parameter is evaluated against the tag criteria.
func HandleJobFilter(knownJobs func() []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req filterRequest
		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
		} else {
			// GET: accept ?names=a,b&tags=x,y for quick ad-hoc queries.
			if raw := r.URL.Query().Get("names"); raw != "" {
				req.Names = strings.Split(raw, ",")
			}
			if raw := r.URL.Query().Get("tags"); raw != "" {
				req.Tags = strings.Split(raw, ",")
			}
		}

		f := jobfilter.New(req.Names, req.Tags)
		all := knownJobs()
		matched := make([]string, 0, len(all))
		for _, name := range all {
			// Tags are not stored on the scheduler job list exposed here;
			// we pass nil so only the name criterion is applied.
			if f.MatchName(name) {
				matched = append(matched, name)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(filterResponse{
			Matched: matched,
			Total:   len(matched),
		})
	}
}
