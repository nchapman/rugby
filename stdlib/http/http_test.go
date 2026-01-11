package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("X-Custom", "test-value")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	}))
	defer server.Close()

	resp, err := Get(server.URL)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if resp.Status != 200 {
		t.Errorf("Status = %d, want 200", resp.Status)
	}
	if !resp.Ok() {
		t.Error("Ok() = false, want true")
	}
	if resp.Body() != "hello world" {
		t.Errorf("Body() = %q, want %q", resp.Body(), "hello world")
	}
	if resp.Headers["X-Custom"] != "test-value" {
		t.Errorf("Headers[X-Custom] = %q, want %q", resp.Headers["X-Custom"], "test-value")
	}
}

func TestGetWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := Get(server.URL, Options{
		Headers: map[string]string{"Authorization": "Bearer test-token"},
	})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	resp, err := Post(server.URL)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp.Status != 201 {
		t.Errorf("Status = %d, want 201", resp.Status)
	}
}

func TestPostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		body, _ := io.ReadAll(r.Body)
		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if data["name"] != "Alice" {
			t.Errorf("body.name = %v, want Alice", data["name"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := Post(server.URL, Options{
		JSON: map[string]any{"name": "Alice", "age": 30},
	})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
}

func TestPostForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", ct)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if r.Form.Get("username") != "alice" {
			t.Errorf("form.username = %q, want alice", r.Form.Get("username"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := Post(server.URL, Options{
		Form: map[string]string{"username": "alice", "password": "secret"},
	})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
}

func TestResponseJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name": "Alice", "age": 30}`))
	}))
	defer server.Close()

	resp, err := Get(server.URL)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	data, err := resp.JSON()
	if err != nil {
		t.Fatalf("JSON() error = %v", err)
	}
	if data["name"] != "Alice" {
		t.Errorf("data.name = %v, want Alice", data["name"])
	}
}

func TestResponseJSONArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[1, 2, 3]`))
	}))
	defer server.Close()

	resp, err := Get(server.URL)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	arr, err := resp.JSONArray()
	if err != nil {
		t.Fatalf("JSONArray() error = %v", err)
	}
	if len(arr) != 3 {
		t.Errorf("len(arr) = %d, want 3", len(arr))
	}
}

func TestPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := Put(server.URL)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
}

func TestPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := Patch(server.URL)
	if err != nil {
		t.Fatalf("Patch() error = %v", err)
	}
}

func TestDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	resp, err := Delete(server.URL)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if resp.Status != 204 {
		t.Errorf("Status = %d, want 204", resp.Status)
	}
}

func TestHead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "HEAD" {
			t.Errorf("expected HEAD, got %s", r.Method)
		}
		w.Header().Set("Content-Length", "1234")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := Head(server.URL)
	if err != nil {
		t.Fatalf("Head() error = %v", err)
	}
	if resp.Headers["Content-Length"] != "1234" {
		t.Errorf("Content-Length = %q, want 1234", resp.Headers["Content-Length"])
	}
}

func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := Get(server.URL, Options{Timeout: 10 * time.Millisecond})
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestOkForVariousStatusCodes(t *testing.T) {
	tests := []struct {
		status int
		want   bool
	}{
		{200, true},
		{201, true},
		{204, true},
		{299, true},
		{300, false},
		{400, false},
		{404, false},
		{500, false},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.status), func(t *testing.T) {
			r := &Response{Status: tt.status}
			if got := r.Ok(); got != tt.want {
				t.Errorf("Ok() for status %d = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
