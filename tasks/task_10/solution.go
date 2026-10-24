package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	tasks := make(map[string]Task, 0)

	return &inMemoryTaskRepo{
		clock: clock,
		tasks: tasks,
	}
}

func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	var task Task

	title = strings.TrimSpace(title)

	if len(title) == 0 {
		return task, ErrInvalidTitle
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	id := rand.Text()

	task = Task{
		ID:        id,
		Title:     title,
		UpdatedAt: r.clock.Now(),
	}
	r.tasks[id] = task

	return task, nil
}

func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	return task, ok
}

func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0)
	for _, task := range r.tasks {
		if task.Done == done {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]

	if !ok {
		return task, ErrNotFound
	}

	if task.Done != done {
		task.Done = done
		task.UpdatedAt = r.clock.Now()
		r.tasks[task.ID] = task
	}

	return task, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	handler := httpHandler{
		repo: repo,
	}

	return &handler
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/tasks" {
		switch r.Method {
		case http.MethodGet:
			h.handleList(w, r)
		case http.MethodPost:
			h.handleCreate(w, r)
		default:
			w.WriteHeader(405)
		}
		return
	}

	if !strings.HasPrefix(path, "/tasks/") {
		w.WriteHeader(404)
		return
	}
	id := strings.TrimPrefix(path, "/tasks/")
	if id == "" || strings.Contains(id, "/") {
		w.WriteHeader(404)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r, id)
	case http.MethodPatch:
		h.handlePatch(w, r, id)
	default:
		w.WriteHeader(405)
	}
}

func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title *string `json:"title"`
	}
	if err := decodeStrictJSON(r.Body, &body); err != nil || body.Title == nil {
		w.WriteHeader(400)
		return
	}

	task, err := h.repo.Create(*body.Title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	writeJSON(w, 201, task)
}

func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)

	if ok {
		writeJSON(w, 200, task)
	} else {
		w.WriteHeader(404)
	}
}

func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	if len(params["done"]) > 1 {
		w.WriteHeader(400)
		return
	}

	done := false
	_, ok := params["done"]

	if ok {
		switch params.Get("done") {
		case "true":
			done = true
		case "false":
		default:
			w.WriteHeader(400)
			return
		}
	}

	tasks := h.repo.List(done)
	if tasks == nil {
		tasks = make([]Task, 0)
	}
	slices.SortFunc(tasks, func(a, b Task) int {
		if cmp := b.UpdatedAt.Compare(a.UpdatedAt); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.ID, b.ID)
	})

	writeJSON(w, 200, tasks)
}

func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Done *bool `json:"done"`
	}
	if err := decodeStrictJSON(r.Body, &body); err != nil || body.Done == nil {
		w.WriteHeader(400)
		return
	}

	task, err := h.repo.SetDone(id, *body.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	writeJSON(w, 200, task)
}

func decodeStrictJSON(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return err
	}

	var m any
	if err := decoder.Decode(&m); err == nil {
		return errors.New("more than one JSON object")
	} else if err != io.EOF {
		return err
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}
