package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Log only — the response header is already sent.
		_ = err
	}
}

// writeError writes a structured JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

// paginate extracts limit/offset from query parameters with sane defaults.
func paginate(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0

	if l := r.URL.Query().Get("limit"); l != "" {
		var lv int
		if _, err := parseIntParam(l, &lv); err == nil && lv > 0 && lv <= 200 {
			limit = lv
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		var ov int
		if _, err := parseIntParam(o, &ov); err == nil && ov >= 0 {
			offset = ov
		}
	}
	return
}

func parseIntParam(s string, out *int) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, err
	}
	*out = n
	return n, nil
}
