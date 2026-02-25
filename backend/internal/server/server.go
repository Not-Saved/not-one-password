package server

import (
	"context"
	"encoding/json"
	"log"
	"main/internal/adapters/middleware"
	"main/internal/bootstrap"
	"main/internal/oapi"
	"net/http"
	"os"
	"path/filepath"

	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	httpSwagger "github.com/swaggo/http-swagger"
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

func (s *Server) RegisterHandlers(handlers *bootstrap.Handlers, services *bootstrap.Services) {
	si := oapi.NewStrictHandler(handlers, nil)
	apiMux := oapi.HandlerFromMux(si, http.NewServeMux())

	spec, err := oapi.GetSwagger()
	if err != nil {
		log.Fatalf("failed to load swagger spec: %v", err)
	}

	handler := nethttpmiddleware.OapiRequestValidatorWithOptions(spec, &nethttpmiddleware.Options{
		ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(oapi.ErrorResponse{
				Code:    statusCode,
				Message: message,
			})
		},
	})(apiMux)
	handler = middleware.AuthenticatedMiddleware(handler, services.UserService)
	handler = middleware.ClientInfoMiddleware(handler)
	s.mux.Handle("/api/", http.StripPrefix("/api", handler))
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

func (s *Server) RegisterSwaggerRoute() {
	s.mux.HandleFunc("/api/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		swagger, _ := oapi.GetSwagger()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(swagger)
	})
	s.mux.Handle("/api/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/api/swagger.json"),
	))
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.port, s.mux)
}

func withReqResContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ctx = context.WithValue(ctx, "httpRequest", r)
		ctx = context.WithValue(ctx, "httpResponseWriter", w)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
