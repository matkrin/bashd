package logger

import (
	"fmt"
	"log/slog"
	"os"
)

var (
	Logger *slog.Logger
	level  = new(slog.LevelVar)
)

func Init(levelStr string, filename string) {
	var l slog.Level
	switch levelStr {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	level.Set(l)

	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic("No log file")
	}

	handler := slog.NewTextHandler(logfile, &slog.HandlerOptions{
		Level: level,
	})

	Logger = slog.New(handler)
}

func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

func Debugf(msg string, args ...any) {
	Logger.Debug(fmt.Sprintf(msg, args...))
}

func Infof(msg string, args ...any) {
	Logger.Info(fmt.Sprintf(msg, args...))
}

func Warnf(msg string, args ...any) {
	Logger.Warn(fmt.Sprintf(msg, args...))
}

func Errorf(msg string, args ...any) {
	Logger.Error(fmt.Sprintf(msg, args...))
}
