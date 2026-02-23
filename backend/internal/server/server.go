package server

import (
	"main/internal/bootstrap"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	port string
	mux  *http.ServeMux
	api  *http.ServeMux
}

func New(port string) *Server {
	return &Server{
		port: port,
		mux:  http.NewServeMux(),
		api:  http.NewServeMux(),
	}
}

func (s *Server) RegisterApiRoutes(h *bootstrap.Handlers) {
	s.api.HandleFunc("GET /users", h.User.ListUsers)
	s.api.HandleFunc("POST /users", h.User.CreateUser)
	s.mux.Handle("/api/", http.StripPrefix("/api", s.api))
}

func (s *Server) RegisterStaticRoute() {
	s.mux.Handle("/assets/",
		http.StripPrefix("/assets/",
			http.FileServer(http.Dir("./public/assets")),
		),
	)
}

func (s *Server) RegisterSpaRoute(filePath string) {
	s.mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If it's requesting a real file in root (like /favicon.ico)
		path := filepath.Join(filePath, r.URL.Path)
		_, err := os.Stat(path)
		if err == nil {
			http.ServeFile(w, r, path)
			return
		}

		// Otherwise serve SPA
		http.ServeFile(w, r, filepath.Join(filePath, "index.html"))
	}))
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.port, s.mux)
}
