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

type Server interface {
	Start()
	StartInfraServer()
	StopInfraServer()
	Stop()
}

type Config struct {
	Port      uint `koanf:"port"`
	InfraPort uint `koanf:"infraPort"`
}

type httpServer struct {
	Port        uint
	InfraPort   uint
	server      *http.Server
	infraServer *http.Server
}

func NewHTTPServer(router *echo.Echo, conf Config) httpServer {
	return httpServer{
		Port:      conf.Port,
		InfraPort: conf.InfraPort,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", defaultHost, conf.Port),
			Handler: router,
		},
	}
}

func NewInfraHTTPServer(router http.Handler, conf Config) httpServer {
	return httpServer{
		InfraPort: conf.InfraPort,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", defaultHost, conf.InfraPort),
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

func (httpServer httpServer) StartInfraServer() {
	const op = "http.StartInfraServer"
	go func() {
		if err := httpServer.server.ListenAndServe(); err != nil {
			logger.Logger.Named(op).Fatal("failed to stater listening InfraHttpServer...", zap.Uint("port", httpServer.InfraPort), zap.Error(err))
		}
	}()
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

func (httpServer httpServer) StopInfraServer() {
	const op = "http.Stop"
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	if err := httpServer.server.Shutdown(ctx); err != nil {
		logger.Logger.Named(op).Fatal("InfraServer forced to shutdown ", zap.Error(err))
	}
}
