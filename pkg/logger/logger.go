package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *slog.Logger
var once sync.Once

type SlogAdapter struct {
	*slog.Logger
}

type Config struct {
	FilePath         string `koanf:"file_path"`
	UseLocalTime     bool   `koanf:"use_local_time"`
	FileMaxSizeInMB  int    `koanf:"file_max_size_in_mb"`
	FileMaxAgeInDays int    `koanf:"file_max_age_in_days"`
}

func Init(cfg Config) {
	once.Do(func() {
		workingDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current working directory: %v", err)
		}
		fileWriter := &lumberjack.Logger{
			Filename:  filepath.Join(workingDir, cfg.FilePath),
			LocalTime: cfg.UseLocalTime,
			MaxSize:   cfg.FileMaxSizeInMB,
			MaxAge:    cfg.FileMaxAgeInDays,
		}
		globalLogger = slog.New(
			slog.NewJSONHandler(io.MultiWriter(fileWriter, os.Stdout), &slog.HandlerOptions{}),
		)
	})
}

func New() SlogAdapter {
	return SlogAdapter{
		Logger: globalLogger,
	}
}

func (l SlogAdapter) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Error(msg, keysAndValues...)
}
