package arias

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/myanimestream/arias/aria2"
	"log"
	"net/http"
	"time"
)

var schemaDecoder = schema.NewDecoder()

type Server struct {
	Router chi.Router
	Config Config

	AriaClient aria2.Client
	Storage    Storage

	tasks map[uuid.UUID]Task
}

func NewServer(config Config) (s Server, err error) {
	ariaClient, err := aria2.Dial(config.Aria2Addr)
	if err != nil {
		return
	}

	storage, err := NewStorageFromType(config.StorageType)
	if err != nil {
		return
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	s = Server{
		Router: r,
		Config: config,

		AriaClient: ariaClient,
		Storage:    storage,

		tasks: make(map[uuid.UUID]Task),
	}

	s.addHandlers()
	return
}

func (s *Server) ListenAndServe() error {
	addr := s.Config.ServerAddr
	log.Printf("Starting to serve on %s\n", addr)
	return http.ListenAndServe(addr, s.Router)
}

func (s *Server) PerformTask(task Task) {
	id := task.GetId()
	s.tasks[id] = task
	go func() {
		_ = task.Perform()
		//delete(s.tasks, id)
	}()
}

func (s *Server) addHandlers() {
	r := s.Router

	r.Get("/download", s.download)
	r.Get("/status", s.status)
}

func jsonResponse(w http.ResponseWriter, data interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func (s *Server) download(w http.ResponseWriter, r *http.Request) {
	var downloadRequest DownloadRequest
	err := schemaDecoder.Decode(&downloadRequest, r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := downloadRequest.Check(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	task := NewDownloadTask(s, downloadRequest)
	s.PerformTask(task)

	resp := DownloadResponse{task.GetId().String()}
	_ = jsonResponse(w, resp, http.StatusOK)
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	rawID := r.URL.Query().Get("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	task, ok := s.tasks[id]
	if !ok {
		http.Error(w, "Task not found", 404)
		return
	}

	_ = jsonResponse(w, task.GetStatus(), http.StatusOK)
}
