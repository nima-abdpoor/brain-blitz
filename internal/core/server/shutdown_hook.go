package server

import (
	"errors"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func addShutdownHook(closers ...io.Closer) {
	zap.L().Info("listening signals...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-c
	zap.L().Info("graceful shutdown...")
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			zap.L().Error("failed to stop closer", zap.Any("err", err))
		}
	}

	zap.L().Info("completed graceful shutdown")
	if err := zap.L().Sync(); err != nil {
		if !errors.Is(err, syscall.ENOTTY) {
			log.Printf("failed to flush logger err=%v\n", err)
		} else {
			log.Printf("failed to flush logger syscall.ENOTTY==> err=%v\n", err)
		}
	}
}
