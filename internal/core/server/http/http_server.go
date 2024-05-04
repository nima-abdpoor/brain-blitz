package http

import (
	"BrainBlitz.com/game/internal/infra/config"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"time"
)

const defaultHost = "0.0.0.0"

type HttpServer interface {
	Start()
	Stop()
}

type httpServer struct {
	Port   uint
	server *http.Server
}

func NewHTTPServer(router *echo.Echo, conf config.HttpServerConfig) httpServer {
	return httpServer{
		Port: conf.Port,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", defaultHost, conf.Port),
			Handler: router,
		},
	}
}

func (httpServer httpServer) Start() {
	go func() {
		if err := httpServer.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf(
				"failed to stater HttpServer listen port %d, err=%s",
				httpServer.Port, err.Error(),
			)
		}
	}()
	log.Printf("Start Service with port %d", httpServer.Port)
}

func (httpServer httpServer) Stop() {
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	if err := httpServer.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown err=%s", err.Error())
	}
}
