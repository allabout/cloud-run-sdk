package util

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	sigCh           = make(chan os.Signal, 2)
)

// SetupSignalHandler registers for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	signal.Notify(sigCh, shutdownSignals...)
	go func() {
		<-sigCh
		close(stop)
		<-sigCh
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

func InjectSignal(sig os.Signal) {
	sigCh <- sig
}
