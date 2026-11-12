package pihole

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCloseAllLogsOut(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(&NewClientInput{BaseURL: srv.URL, Password: "test"})
	c.sid = "test"

	CloseAll(context.Background())

	if len(paths) != 1 || paths[0] != "DELETE /api/auth" {
		t.Fatalf("requests = %v, want [DELETE /api/auth]", paths)
	}
	if c.sid != "" {
		t.Fatalf("sid = %q, want empty", c.sid)
	}
}
