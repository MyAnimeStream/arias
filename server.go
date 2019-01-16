package arias

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/MyAnimeStream/arias/aria2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"time"
)

const Version = "0.1.1"

var schemaDecoder = schema.NewDecoder()

type Server struct {
	Router     chi.Router
	HttpClient *http.Client
	Config     Config

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

	//r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	s = Server{
		Router:     r,
		HttpClient: &http.Client{Timeout: 30 * time.Second},
		Config:     config,

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
	}()
}

func (s *Server) SendCallback(url string, data interface{}) (resp *http.Response, err error) {
	p, err := json.Marshal(data)
	if err != nil {
		return
	}
	w := bytes.NewBuffer(p)

	req, err := http.NewRequest("POST", url, w)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("arias/%s", Version))

	resp, err = s.HttpClient.Do(req)
	return
}

func (s *Server) GoSendCallback(url string, data interface{}) {
	go func() {
		_, err := s.SendCallback(url, data)
		if err != nil {
			log.Printf("couldn't perform callback to %s: %s\n", url, err)
		}
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

	if err := downloadRequest.UseConfig(&s.Config); err != nil {
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
		http.Error(w, "task not found", 404)
		return
	}

	_ = jsonResponse(w, task.GetStatus(), http.StatusOK)
}
