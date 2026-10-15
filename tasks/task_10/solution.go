package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
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
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{clock: clock, tasks: make(map[string]Task)}
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return Task{}, ErrInvalidTitle
	}

	r.seq++
	temp := strconv.FormatUint(r.seq, 10)
	task := Task{
		ID:        temp,
		Title:     trimmedTitle,
		Done:      false,
		UpdatedAt: r.clock.Now()}
	r.tasks[temp] = task

	return task, nil
}
func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.tasks[id]; !ok {
		return Task{}, false
	}
	return r.tasks[id], true
}
func (r *inMemoryTaskRepo) List(done bool) []Task {
	newList := []Task{}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.tasks {
		if v.Done == done {
			newList = append(newList, v)
		}
	}

	return newList
}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[id]; !ok {
		return Task{}, ErrNotFound
	}
	if done != r.tasks[id].Done {
		newTask := r.tasks[id]
		newTask.Done = done
		newTask.UpdatedAt = r.clock.Now()
		r.tasks[id] = newTask
	}

	return r.tasks[id], nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{repo: repo}
}
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/tasks":
		switch r.Method {
		case http.MethodPost:
			h.handleCreate(w, r)
		case http.MethodGet:
			h.handleList(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case strings.HasPrefix(r.URL.Path, "/tasks/"):
		id := strings.TrimPrefix(r.URL.Path, "/tasks/")

		if id != "" && !strings.Contains(id, "/") {
			switch r.Method {
			case http.MethodGet:
				h.handleGet(w, r, id)
			case http.MethodPatch:
				h.handlePatch(w, r, id)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title *string `json:"title"`
	}
	if err := decodeStrictJSON(r.Body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Title == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(*req.Title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	writeJSON(w, http.StatusCreated, task)
}
func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, task)
}
func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	done := false
	if values, ok := r.URL.Query()["done"]; ok {
		if len(values) != 1 || (values[0] != "true" && values[0] != "false") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		done = values[0] == "true"
	}

	tasks := h.repo.List(done)
	if tasks == nil {
		tasks = []Task{}
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].UpdatedAt.Equal(tasks[j].UpdatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})
	writeJSON(w, http.StatusOK, tasks)
}
func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Done *bool `json:"done"`
	}
	if err := decodeStrictJSON(r.Body, &req); err != nil || req.Done == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *req.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	writeJSON(w, http.StatusOK, task)
}
func decodeStrictJSON(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
