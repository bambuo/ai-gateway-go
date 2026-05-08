package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	l, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	log = l.Sugar()
}

func Setup(level string) {
	var lvl zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = zapcore.DebugLevel
	case "info":
		lvl = zapcore.InfoLevel
	case "warn":
		lvl = zapcore.WarnLevel
	case "error":
		lvl = zapcore.ErrorLevel
	default:
		lvl = zapcore.InfoLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	l, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	log = l.Sugar()
}

func Debug(msg string, args ...any) { log.Debugw(msg, args...) }
func Info(msg string, args ...any)  { log.Infow(msg, args...) }
func Warn(msg string, args ...any)  { log.Warnw(msg, args...) }
func Error(msg string, args ...any) { log.Errorw(msg, args...) }

func Audit(clientName, method, path string, status int) {
	log.Infow("请求",
		"client", clientName,
		"method", method,
		"path", path,
		"status", status,
	)
}
