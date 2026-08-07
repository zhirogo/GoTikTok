// Package main 是 GoTikTok 后端服务进程的入口，负责启动 HTTP 服务与优雅停机。
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zhirogo/gotiktok/interface/httpapi"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "服务启动失败: %v\n", err)

		os.Exit(1)
	}
}

// run 组装依赖、启动 HTTP 服务，并在收到退出信号后执行优雅停机。
func run() error {
	gin.SetMode(gin.ReleaseMode)

	logger, err := newLogger()
	if err != nil {
		return fmt.Errorf("初始化日志: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	addr := os.Getenv("GOTIKTOK_HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewRouter(logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		logger.Info("HTTP 服务启动", zap.String("addr", addr))

		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("HTTP 服务异常退出: %w", err)
	case <-ctx.Done():
	}

	logger.Info("收到退出信号，开始优雅停机")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("优雅停机失败: %w", err)
	}

	logger.Info("服务已退出")

	return nil
}

// newLogger 构建输出到 stdout 的 zap 结构化 JSON 日志。
func newLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	return cfg.Build()
}
