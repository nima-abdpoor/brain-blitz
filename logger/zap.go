package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"sync"
)

var once = sync.Once{}
var Logger *zap.Logger

func init() {
	once.Do(func() {
		zapLogger, err := zap.NewProduction()
		if err != nil {
			fmt.Println("Error initializing logger")
		}

		config := zap.NewProductionEncoderConfig()
		config.EncodeTime = zapcore.ISO8601TimeEncoder
		defualtEncoder := zapcore.NewJSONEncoder(config)

		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:  "./logs/log.json",
			LocalTime: false,
			MaxSize:   10, // megabytes
			//MaxBackups: 10,
			MaxAge: 30, // days
		})
		stdoutWriter := zapcore.AddSync(os.Stdout)
		defaultLogLevel := zapcore.InfoLevel
		core := zapcore.NewTee(
			zapcore.NewCore(defualtEncoder, writer, defaultLogLevel),
			zapcore.NewCore(defualtEncoder, stdoutWriter, defaultLogLevel),
		)
		zapLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		Logger = zapLogger
	})
}
