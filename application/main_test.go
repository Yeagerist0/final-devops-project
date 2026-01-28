package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	store, err := newStore(path)
	if err != nil {
		t.Fatal(err)
	}
	app := &App{store: store, token: "test-token", greeting: "Hello"}
	h := app.handler()
	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if got := request("GET", "/healthz", "", "").Code; got != http.StatusOK {
		t.Fatalf("health: %d", got)
	}
	if got := request("POST", "/api/tasks", "wrong", `{"title":"first"}`).Code; got != http.StatusUnauthorized {
		t.Fatalf("auth: %d", got)
	}
	if got := request("POST", "/api/tasks", "test-token", `{"title":" "}`).Code; got != http.StatusBadRequest {
		t.Fatalf("validation: %d", got)
	}
	if got := request("POST", "/api/tasks", "test-token", `{"title":"first"}`).Code; got != http.StatusCreated {
		t.Fatalf("create: %d", got)
	}
	loaded, err := newStore(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := loaded.list()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "first" {
		t.Fatalf("persisted tasks: %+v", got)
	}
	response := request("GET", "/api/tasks", "", "")
	var result struct {
		Tasks []Task `json:"tasks"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Tasks) != 1 {
		t.Fatalf("tasks: %+v", result.Tasks)
	}
	if body := request("GET", "/metrics", "", "").Body.String(); !strings.Contains(body, "app_tasks_stored 1") {
		t.Fatal(body)
	}
	if got := request("PUT", "/api/tasks/1", "test-token", `{"title":"updated","done":true}`).Code; got != http.StatusOK {
		t.Fatalf("update: %d", got)
	}
	if body := request("GET", "/api/tasks/stats", "", "").Body.String(); !strings.Contains(body, `"completed":1`) {
		t.Fatal(body)
	}
	if got := request("DELETE", "/api/tasks/1", "test-token", "").Code; got != http.StatusNoContent {
		t.Fatalf("delete: %d", got)
	}
	if got := request("DELETE", "/api/tasks/1", "test-token", "").Code; got != http.StatusNotFound {
		t.Fatalf("missing delete: %d", got)
	}
}
