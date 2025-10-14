package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"github.com/redis/go-redis/v9"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"google.golang.org/api/option"

	"chatbot/config"
	v1 "chatbot/internal/controller/http"
	"chatbot/internal/usecase"

	"chatbot/pkg/httpserver"
	"chatbot/pkg/postgres"
)

func Run(cfg *config.Config) {

	loc, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		panic(err)
	}

	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// "time" atributini Asia/Tashkentga o'zgartirish
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.Attr{
						Key:   slog.TimeKey,
						Value: slog.AnyValue(t.In(loc)),
					}
				}
			}
			return a
		},
	}))

	slog.SetDefault(slogger)

	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
	}
	defer pg.Close()

	// Use case
	useCase := usecase.New(pg, cfg)

	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--disable-gpu",
			"--no-sandbox",
			"--headless",
		},
	}
	caps.AddChrome(chromeCaps)

	// Gemini
	ctx := context.Background()
	gemini_client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.ApiKey.Key))
	if err != nil {
		slog.Error("failed to create Gemini client", "error", err)
	}
	defer gemini_client.Close()

	// redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "4545",
		DB:       0,
	})

	// HTTP Server
	handler := gin.New()
	v1.NewRouter(handler, cfg, useCase, gemini_client, rdb)

	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	slog.Info("app - Run - httpServer: %s", cfg.HTTP.Port)

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		slog.Info("app - Run - signal: %s", s.String())
	case err = <-httpServer.Notify():
		slog.Error("app - Run - httpServer.Notify:", err)
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		slog.Error("app - Run - httpServer.Shutdown: %w", err)
	}

}
