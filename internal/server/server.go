package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wishmatic/elfu-mcp/internal/auth"
	"github.com/wishmatic/elfu-mcp/internal/config"
	"github.com/wishmatic/elfu-mcp/internal/filestore"
	mcpServer "github.com/wishmatic/elfu-mcp/internal/mcp"
	"github.com/wishmatic/elfu-mcp/internal/resolve"
	"go.uber.org/zap"
)

const writeTimeout = 10 * time.Minute

type Server struct {
	cfg    config.Config
	log    *zap.Logger
	files  *filestore.Client
	router *chi.Mux
	http   *http.Server
}

func New(cfg config.Config, log *zap.Logger) (*Server, error) {
	if cfg.APIKey == "" {
		return nil, auth.ErrNoAPIKey
	}

	publicBase, err := cfg.PublicBase()
	if err != nil {
		return nil, err
	}

	if publicBase == nil {
		return nil, fmt.Errorf("PUBLIC_HOST is required")
	}

	if cfg.FilesDir == "" {
		return nil, fmt.Errorf("FILES_DIR must not be empty")
	}

	files, err := filestore.New(filestore.Config{
		Dir:        cfg.FilesDir,
		PublicBase: publicBase,
	}, log)
	if err != nil {
		return nil, fmt.Errorf("configure file storage: %w", err)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.ClientIPFromRemoteAddr)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	files.Register(router)

	log.Info("local files enabled", zap.String("dir", cfg.FilesDir))
	log.Warn("stored files are readable by anyone with the URL")

	resolver, err := resolve.New(files, publicBase.String())
	if err != nil {
		return nil, fmt.Errorf("build image resolver: %w", err)
	}

	mcpSrv, err := mcpServer.New(mcpServer.Deps{
		Log:      log,
		Resolver: resolver,
	})
	if err != nil {
		return nil, fmt.Errorf("build mcp server: %w", err)
	}

	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpSrv
	}, nil)

	protected := auth.Middleware(log, cfg.APIKey)(mcpHandler)

	router.Mount("/mcp", protected)
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &Server{
		cfg:    cfg,
		log:    log,
		files:  files,
		router: router,
		http: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       60 * time.Second,
		},
	}, nil
}

func (s *Server) Run() error {
	s.log.Info("server listening", zap.String("addr", s.cfg.Addr()))

	if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
