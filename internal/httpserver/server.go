package httpserver

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nikitanaumenko/calendar/internal/api/gen"
	"go.uber.org/zap"
)

func NewRouter(log *zap.Logger, api gen.StrictServerInterface, staticDir string) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(Logger(log))
	router.Use(Recoverer(log))
	router.Use(middleware.NoCache)

	strict := gen.NewStrictHandlerWithOptions(api, nil, gen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		},
	})
	gen.HandlerFromMux(strict, router)
	gen.HandlerFromMuxWithBaseURL(strict, router, "/api")

	if staticDir != "" {
		mountStatic(router, staticDir)
	}

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusNotFound, "not_found", "resource not found")
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	})

	return router
}

func mountStatic(router chi.Router, staticDir string) {
	fileServer := http.FileServer(http.Dir(staticDir))
	router.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})
	router.Get("/", serveIndex(staticDir))
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			WriteError(w, http.StatusNotFound, "not_found", "resource not found")
			return
		}
		serveIndex(staticDir)(w, r)
	})
}

func serveIndex(staticDir string) http.HandlerFunc {
	indexPath := filepath.Join(staticDir, "index.html")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}
		if _, err := os.Stat(indexPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				WriteError(w, http.StatusNotFound, "not_found", "frontend build not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			return
		}
		http.ServeFile(w, r, indexPath)
	}
}
