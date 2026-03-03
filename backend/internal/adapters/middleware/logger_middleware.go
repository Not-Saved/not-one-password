package middleware

import (
	"log"
	"net/http"
	"time"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		startTime := time.Now()

		next.ServeHTTP(writer, req)

		elapsedTime := time.Since(startTime)
		log.Printf("[%s] [%s] [%s]\n", req.Method, req.URL.Path, elapsedTime)
	})
}
