package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer test-token")
		}
		if r.Method != http.MethodGet {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/v1/pages" {
			t.Errorf("Path = %q, want /v1/pages", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "pg_123"})
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	var result map[string]string
	err := client.Get(context.Background(), "/pages", &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result["id"] != "pg_123" {
		t.Errorf("id = %q, want %q", result["id"], "pg_123")
	}
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Test Page" {
			t.Errorf("name = %q, want %q", body["name"], "Test Page")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "pg_new"})
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	reqBody := map[string]string{"name": "Test Page"}
	var result map[string]string
	err := client.Post(context.Background(), "/pages", reqBody, &result)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if result["id"] != "pg_new" {
		t.Errorf("id = %q, want %q", result["id"], "pg_new")
	}
}

func TestClientError401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient("bad-token", server.URL)
	var result map[string]string
	err := client.Get(context.Background(), "/pages", &result)
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
}

func TestClientError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	var result map[string]string
	err := client.Get(context.Background(), "/pages/nonexistent", &result)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func TestClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	err := client.Delete(context.Background(), "/pages/pg_123")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestClientPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Method = %q, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "pg_123", "name": "Updated"})
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	reqBody := map[string]string{"name": "Updated"}
	var result map[string]string
	err := client.Put(context.Background(), "/pages/pg_123", reqBody, &result)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if result["name"] != "Updated" {
		t.Errorf("name = %q, want %q", result["name"], "Updated")
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "pg_123"})
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var result map[string]string
	err := client.Get(ctx, "/pages", &result)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
