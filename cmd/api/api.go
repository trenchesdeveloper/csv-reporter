package main

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"net/http"
	"time"

	"github.com/trenchesdeveloper/csv-reporter/config"
	db "github.com/trenchesdeveloper/csv-reporter/db/sqlc"
)

type server struct {
	config *config.AppConfig
	store  db.Store
	logger *zap.SugaredLogger
}

func (s *server) mount() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(middleware.Timeout(60 * time.Second))
	r.Route("/v1", func(r chi.Router) {
		//r.Get("/health", s.healthCheckHandler)
		docsURL := fmt.Sprintf("%s/swagger/doc.json", s.config.SERVER_PORT)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsURL), //The url pointing to API definition
		))
	})

	r.Route("/users", func(r chi.Router) {

	})

	// Public routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", s.SignupHandler)
		// r.Post("/login", s.LoginHandler)
	})
	return r
}

func (s *server) start(mux http.Handler) error {

	srv := &http.Server{
		Addr:         s.config.SERVER_PORT,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	s.logger.Infow("starting server", "port", s.config.SERVER_PORT)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Errorw("Server failed to start", "error", err)
	}

	return nil
}
