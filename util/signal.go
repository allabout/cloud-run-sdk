package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var (
	stopCh               = make(chan os.Signal, 2)
	onlyOneSignalHandler = make(chan struct{})
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
)

// SetupSignalHandler registers for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler() context.Context {
	close(onlyOneSignalHandler) // panics when called twice

	ctx, cancel := context.WithCancel(context.Background())

	signal.Notify(stopCh, shutdownSignals...)
	go func() {
		<-stopCh
		cancel()
		<-stopCh
		os.Exit(1) // second signal. Exit directly.
	}()

	return ctx
}
