package server

import (
	"main/internal/bootstrap"
	"main/internal/docs"
	"main/internal/oapi"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	port string
	mux  *http.ServeMux
}

func New(port string) *Server {
	return &Server{
		port: port,
		mux:  http.NewServeMux(),
	}
}

func (s *Server) RegisterHandlers(handlers *bootstrap.Handlers) {
	apiMux := oapi.HandlerFromMux(handlers, http.NewServeMux())
	s.mux.Handle("/api/", http.StripPrefix("/api", apiMux))
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

func (s *Server) RegisterOpenapiRouter() {
	s.mux.HandleFunc("/api/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(docs.OpenAPISpec)
	})
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.port, s.mux)
}
