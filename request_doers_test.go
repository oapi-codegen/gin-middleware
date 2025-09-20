package ginmiddleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func doGet(t *testing.T, handler http.Handler, rawURL string) *httptest.ResponseRecorder {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid url: %s", rawURL)
	}

	r, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		t.Fatalf("Could not construct a request: %s", rawURL)
	}
	r.Header.Set("accept", "application/json")
	r.Header.Set("host", u.Host)

	tt := httptest.NewRecorder()

	handler.ServeHTTP(tt, r)

	return tt
}

func doPost(t *testing.T, handler http.Handler, rawURL string, jsonBody interface{}) *httptest.ResponseRecorder {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid url: %s", rawURL)
	}

	body, err := json.Marshal(jsonBody)
	if err != nil {
		t.Fatalf("Could not marshal request body: %v", err)
	}

	r, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Could not construct a request for URL %s: %v", rawURL, err)
	}
	r.Header.Set("accept", "application/json")
	r.Header.Set("content-type", "application/json")
	r.Header.Set("host", u.Host)

	tt := httptest.NewRecorder()

	handler.ServeHTTP(tt, r)

	return tt
}
