package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/exp/slog"
)

type Orchestrator struct {
	Port    int
	Nodes   map[string]*NodeConnection
	LogFile string
}

func (orchestrator *Orchestrator) Run() (err error) {
	if orchestrator.Nodes == nil {
		orchestrator.Nodes = make(map[string]*NodeConnection)
	}
	var logger *slog.Logger
	var closeFn func()
	if logger, closeFn, err = setupLogging(orchestrator.LogFile); err != nil {
		return err
	}
	defer closeFn()
	handleTermination(logger)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	done := func() {
		cancel()
		wg.Done()
	}

	wg.Add(1)
	go OServerThread(orchestrator, ctx, logger, done)
	wg.Add(1)
	go DisconnectThread(orchestrator, ctx, logger, done)

	logger.Info("Orchestrator is up and running!")
	wg.Wait()
	cancel()
	return nil
}

func setupLogging(destination string) (*slog.Logger, func(), error) {
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

	textHandler := slog.NewTextHandler(logDest)
	logger := slog.New(textHandler)
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
