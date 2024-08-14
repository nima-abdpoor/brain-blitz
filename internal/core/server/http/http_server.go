package http

import (
	"BrainBlitz.com/game/logger"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const defaultHost = "0.0.0.0"

type HttpServer interface {
	Start()
	Stop()
}

type Config struct {
	Port uint
}

type httpServer struct {
	Port   uint
	server *http.Server
}

func NewHTTPServer(router *echo.Echo, conf Config) httpServer {
	return httpServer{
		Port: conf.Port,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", defaultHost, conf.Port),
			Handler: router,
		},
	}
}

func (httpServer httpServer) Start() {
	const op = "http.Start"
	go func() {
		if err := httpServer.server.ListenAndServe(); err != nil {
			logger.Logger.Named(op).Fatal("failed to stater listening HttpServer...", zap.Uint("port", httpServer.Port), zap.Error(err))
		}
	}()
	logger.Logger.Named(op).Info("http server started", zap.Uint("port", httpServer.Port))
}

func (httpServer httpServer) Stop() {
	const op = "http.Stop"
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	if err := httpServer.server.Shutdown(ctx); err != nil {
		logger.Logger.Named(op).Fatal("Server forced to shutdown", zap.Error(err))
	}
}
