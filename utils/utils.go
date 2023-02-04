package utils

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/exp/slog"
)

func SetupLogging(destination string) (*slog.Logger, func(), error) {
	var logDest io.Writer
	var file *os.File
	if destination == "" {
		logDest = os.Stdout
	} else {
		file, err := os.OpenFile(destination, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			slog.Error("Error opening log file:", err)
			return nil, nil, err
		}
		logDest = file
	}

	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	textHandler := opts.NewTextHandler(logDest)
	logger := slog.New(textHandler)

	handleTermination(logger)
	slog.SetDefault(logger)
	return logger, func() { file.Close() }, nil
}

func handleTermination(logger *slog.Logger) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigc
		logger.Error(fmt.Sprintf("Received signal: %s", sig.String()), nil)
		os.Exit(1)
	}()
}
