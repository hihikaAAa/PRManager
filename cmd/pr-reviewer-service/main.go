package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"

	"github.com/hihikaAAa/PRManager/internal/config"
	pullrequesthandlercreate "github.com/hihikaAAa/PRManager/internal/http-server/handlers/pullrequest/create"
	pullrequesthandlersmerge "github.com/hihikaAAa/PRManager/internal/http-server/handlers/pullrequest/merge"
	pullrequesthandlerreassign "github.com/hihikaAAa/PRManager/internal/http-server/handlers/pullrequest/reassign"
	teamhandleradd "github.com/hihikaAAa/PRManager/internal/http-server/handlers/team/add"
	teamhandlerget "github.com/hihikaAAa/PRManager/internal/http-server/handlers/team/get"
	userhandlergetreview "github.com/hihikaAAa/PRManager/internal/http-server/handlers/user/getReview"
	userhandlerisactive "github.com/hihikaAAa/PRManager/internal/http-server/handlers/user/isActive"
	statsservice "github.com/hihikaAAa/PRManager/internal/services/statsservice"
    statshandler "github.com/hihikaAAa/PRManager/internal/http-server/handlers/stats/getStats"
	mwlogger "github.com/hihikaAAa/PRManager/internal/http-server/middleware/logger"
	slogpretty "github.com/hihikaAAa/PRManager/internal/lib/logger/slogpretty"
	"github.com/hihikaAAa/PRManager/internal/lib/logger/sl"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres"
	"github.com/hihikaAAa/PRManager/internal/services/prservice"
	"github.com/hihikaAAa/PRManager/internal/services/teamservice"
	"github.com/hihikaAAa/PRManager/internal/services/userservice"
	"github.com/hihikaAAa/PRManager/internal/storage"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	_ = godotenv.Load("local.env")

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	db, err := storage.New(cfg.DB.DSN)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	defer db.Close()

	prRepo := postgres.New(db)
	userRepo := postgres.NewUserRepository(db)
	teamRepo := postgres.NewTeamRepository(db)

	prService := prservice.New(prRepo, userRepo)
	teamService := teamservice.New(userRepo, teamRepo)
	userService := userservice.New(prRepo, userRepo)
	statService := statsservice.New(prRepo)


	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	router.Route("/team", func(r chi.Router) {
		r.Post("/add", teamhandleradd.New(log, teamService))
		r.Get("/get", teamhandlerget.New(log, teamService))
	})

	router.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userhandlerisactive.New(log, userService))
		r.Get("/getReview", userhandlergetreview.New(log, userService))
	})

	router.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", pullrequesthandlercreate.New(log, prService))
		r.Post("/merge", pullrequesthandlersmerge.New(log, prService))
		r.Post("/reassign", pullrequesthandlerreassign.New(log, prService))
	})

	router.Get("/stats", statshandler.New(log, statService))

	srv := &http.Server{
		Addr: cfg.HTTPServer.Address,      
		Handler: router,
		ReadTimeout: cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout: cfg.HTTPServer.IdleTimeout,
	}

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", sl.Err(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", sl.Err(err))
	} else {
		log.Info("server gracefully stopped")
	}
}


func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
		case envLocal:
			log = setupPrettySlog()
		case envDev:
			log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
		case envProd:
			log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
		default:
			log = setupPrettySlog()
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
