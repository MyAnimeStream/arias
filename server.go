package arias

import "net/http"

type Server struct {
	config Config
	mux    *http.ServeMux
}

func NewServer(config Config) Server {
	mux := http.NewServeMux()
	server := Server{config, mux}
	server.addHandlers()

	return server
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.config.ServerAddr, s.mux)
}

func (s *Server) addHandlers() {
	m := s.mux

	m.HandleFunc("/download", s.download)
}

func (s *Server) download(w http.ResponseWriter, r *http.Request) {

}
