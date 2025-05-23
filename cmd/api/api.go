package main

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/trenchesdeveloper/csv-reporter/helpers"
	"go.uber.org/zap"

	"net/http"
	"time"

	"github.com/trenchesdeveloper/csv-reporter/config"
	db "github.com/trenchesdeveloper/csv-reporter/db/sqlc"
)

type server struct {
	config          *config.AppConfig
	store           db.Store
	logger          *zap.SugaredLogger
	tokenManager    *helpers.JwtManager
	sqsClient       *sqs.Client
	presignedClient *s3.PresignClient
}

func (s *server) mount() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(middleware.Timeout(60 * time.Second))
	r.Route("/api/v1", func(r chi.Router) {
		//r.Get("/health", s.healthCheckHandler)
		docsURL := fmt.Sprintf("%s/swagger/doc.json", s.config.SERVER_PORT)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsURL), //The url pointing to API definition
		))

		// ping route
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("pong"))
		})

		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", s.SignupHandler)
			r.Post("/login", s.SigninHandler)
			r.Post("/refresh", s.RefreshTokenHandler)
		})

		//reports route
		r.Route("/reports", func(r chi.Router) {
			r.Use(NewAuthMiddleware(s.tokenManager, s.store))
			r.Post("/", s.CreateReportHandler)
			r.Get("/{reportId}", s.GetReportHandler)
		})
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
