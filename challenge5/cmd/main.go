package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minherz/aichallenges/challenge1/pkg/agents"
	"github.com/minherz/aichallenges/challenge1/pkg/utils"
)

func setupLogging() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(group []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.LevelKey:
				a.Key = "severity"
				if level := a.Value.Any().(slog.Level); level == slog.LevelWarn {
					a.Value = slog.StringValue("WARNING")
				}
			case slog.MessageKey:
				a.Key = "message"
			case slog.TimeKey:
				a.Key = "timestamp"
			}
			return a
		},
	}
	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(jsonHandler))
}

func main() {
	setupLogging()
	e := echo.New()
	if os.Getenv("DO_DEBUG") != "" {
		e.Use(middleware.Logger())
	}
	e.Use(
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
		}),
		middleware.StaticWithConfig(middleware.StaticConfig{
			Root:  "web/static",
			HTML5: true,
		}),
		middleware.GzipWithConfig(middleware.GzipConfig{
			Level: 5,
		}),
		middleware.Secure(),
	)
	e.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(false),   // e.g. ipv4 start with 127.
		echo.TrustLinkLocal(false),  // e.g. ipv4 start with 169.254
		echo.TrustPrivateNet(false), // e.g. ipv4 start with 10. or 192.168
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	agent, err := agents.NewRagAgent(ctx)
	if err != nil {
		e.Logger.Fatal("failed to initialize RAG agent: %q", err.Error())
	}
	e.POST("/ask", agent.Handler)
	// start server
	go func() {
		port := utils.GetEnvOrDefault("PORT", "8080")
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()
	// wait for interrupt signal and gracefully shutdown the server after 5 seconds.
	<-ctx.Done()
	agent.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
