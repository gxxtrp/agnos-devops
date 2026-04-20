package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const aliveFile = "/tmp/worker-alive"

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	endpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("worker"),
			semconv.DeploymentEnvironment(getEnv("ENV", "local")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func touchAliveFile() {
	f, err := os.Create(aliveFile)
	if err != nil {
		slog.Warn("failed to touch alive file", "error", err)
		return
	}
	f.Close()
}

func runJob(ctx context.Context) {
	tracer := otel.Tracer("worker")
	_, span := tracer.Start(ctx, "update-timestamp-job")
	defer span.End()

	slog.Info("updating timestamp for today records",
		"env", getEnv("ENV", "local"),
		"db_host", getEnv("DB_HOST", ""),
		"db_name", getEnv("DB_NAME", ""),
		"timestamp", time.Now().UTC().Format(time.RFC3339),
	)

	// Touch alive file so liveness probe passes
	touchAliveFile()
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := initTracer(ctx)
	if err != nil {
		slog.Warn("failed to initialize tracer, continuing without tracing", "error", err)
	} else {
		defer func() {
			if err := tp.Shutdown(ctx); err != nil {
				slog.Error("failed to shutdown tracer", "error", err)
			}
		}()
	}

	intervalStr := getEnv("WORKER_INTERVAL", "60")
	intervalSec, err := strconv.Atoi(intervalStr)
	if err != nil {
		slog.Warn("invalid WORKER_INTERVAL, using default 60s", "value", intervalStr)
		intervalSec = 60
	}
	interval := time.Duration(intervalSec) * time.Second

	slog.Info("worker starting",
		"env", getEnv("ENV", "local"),
		"interval", interval.String(),
		"db_host", getEnv("DB_HOST", ""),
	)

	// Touch alive file on startup
	touchAliveFile()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			runJob(ctx)
		case sig := <-quit:
			slog.Info("received signal, shutting down", "signal", sig.String())
			return
		}
	}
}
