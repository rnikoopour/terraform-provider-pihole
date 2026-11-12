package pihole

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteDomainEscapesReservedCharacters(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.RequestURI
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(&NewClientInput{BaseURL: srv.URL, Password: "test"})
	c.sid = "test"

	err := c.DeleteDomain(&DeleteDomainInput{
		Ctx:    context.Background(),
		Type:   "allow",
		Kind:   "regex",
		Domain: `(\.|^)sts\.[a-z0-9-]+\.amazonaws\.com$`,
	})
	if err != nil {
		t.Fatalf("DeleteDomain: %v", err)
	}

	want := "/api/domains/allow/regex/%28%5C.%7C%5E%29sts%5C.%5Ba-z0-9-%5D%2B%5C.amazonaws%5C.com%24"
	if got != want {
		t.Fatalf("request URI = %s, want %s", got, want)
	}
}

func TestEscapePathSegment(t *testing.T) {
	tests := map[string]string{
		"fcm.googleapis.com":            "fcm.googleapis.com",
		`(\.|^)pllsadrp\.net$`:          "%28%5C.%7C%5E%29pllsadrp%5C.net%24",
		`10.0.10.240`:                   "10.0.10.240",
		`a+b&c=d:e@f;g,h`:               "a%2Bb%26c%3Dd%3Ae%40f%3Bg%2Ch",
		`https://example.com/hosts.txt`: "https%3A%2F%2Fexample.com%2Fhosts.txt",
	}
	for in, want := range tests {
		if got := escapePathSegment(in); got != want {
			t.Errorf("escapePathSegment(%q) = %q, want %q", in, got, want)
		}
	}
}
