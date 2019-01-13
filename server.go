package arias

import (
	"fmt"
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
		Router:     r,
		Config:     config,
		Storage:    storage,
		AriaClient: ariaClient,
	}

	s.addHandlers()
	return
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.Config.ServerAddr, s.Router)
}

func (s *Server) PerformTask(task Task) {
	s.tasks[task.GetId()] = task
	go func() {
		_ = task.Perform()
	}()
}

func (s *Server) addHandlers() {
	r := s.Router

	r.Get("/download", s.download)
	r.Get("/status", s.status)
}

func (s *Server) download(w http.ResponseWriter, r *http.Request) {
	downloadRequest := defaultDownloadRequest()
	err := schemaDecoder.Decode(&downloadRequest, r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	task := NewDownloadTask(s, downloadRequest)
	s.PerformTask(task)

	_, err = fmt.Fprint(w, task.GetId().URN())
	if err != nil {
		log.Println(err)
	}
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
}
