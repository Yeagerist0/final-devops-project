package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Task struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Done    bool      `json:"done"`
	Created time.Time `json:"created"`
}

type Store struct {
	mu    sync.Mutex
	path  string
	tasks []Task
}

func newStore(path string) (*Store, error) {
	s := &Store{path: path, tasks: []Task{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.tasks); err != nil {
		return nil, fmt.Errorf("read task data: %w", err)
	}
	return s, nil
}

func (s *Store) list() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Task{}, s.tasks...)
}

func (s *Store) add(title string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := 1
	for _, task := range s.tasks {
		if task.ID >= id {
			id = task.ID + 1
		}
	}
	task := Task{ID: id, Title: title, Created: time.Now().UTC()}
	next := append(append([]Task{}, s.tasks...), task)
	if err := s.save(next); err != nil {
		return Task{}, err
	}
	s.tasks = next
	return task, nil
}

func (s *Store) save(tasks []Task) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".tasks-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if err := tmp.Chmod(0600); err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}

type App struct {
	store    *Store
	token    string
	greeting string
	requests atomic.Uint64
	created  atomic.Uint64
}

func (a *App) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", a.home)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); fmt.Fprint(w, "ok\n") })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); fmt.Fprint(w, "ready\n") })
	mux.HandleFunc("GET /metrics", a.metrics)
	mux.HandleFunc("GET /api/tasks", a.listTasks)
	mux.HandleFunc("POST /api/tasks", a.createTask)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.requests.Add(1)
		mux.ServeHTTP(w, r)
	})
}

func (a *App) home(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

func (a *App) listTasks(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"greeting": a.greeting, "tasks": a.store.list()})
}

func (a *App) createTask(w http.ResponseWriter, r *http.Request) {
	provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if a.token == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(a.token)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var input struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	if len(input.Title) == 0 || len(input.Title) > 120 {
		http.Error(w, "title must contain 1-120 characters", http.StatusBadRequest)
		return
	}
	task, err := a.store.add(input.Title)
	if err != nil {
		log.Printf("save task: %v", err)
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	a.created.Add(1)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (a *App) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprint(w, "# HELP app_http_requests_total Total HTTP requests.\n# TYPE app_http_requests_total counter\n")
	fmt.Fprintln(w, "app_http_requests_total", a.requests.Load())
	fmt.Fprint(w, "# HELP app_tasks_created_total Tasks created since startup.\n# TYPE app_tasks_created_total counter\n")
	fmt.Fprintln(w, "app_tasks_created_total", a.created.Load())
	fmt.Fprint(w, "# HELP app_tasks_stored Current persisted tasks.\n# TYPE app_tasks_stored gauge\n")
	fmt.Fprintln(w, "app_tasks_stored", len(a.store.list()))
}

func main() {
	path := getenv("DATA_FILE", "./data/tasks.json")
	store, err := newStore(path)
	if err != nil {
		log.Fatal(err)
	}
	port := getenv("PORT", "8080")
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatal("PORT must be numeric")
	}
	app := &App{store: store, token: os.Getenv("API_TOKEN"), greeting: getenv("GREETING", "Session 21 DevOps Lab")}
	server := &http.Server{Addr: ":" + port, Handler: app.handler(), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
