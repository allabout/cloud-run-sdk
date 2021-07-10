package util

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSetupSignalHandler(t *testing.T) {
	ctx := SetupSignalHandler()

	go func(stopCh chan os.Signal) {
		go handle()
		sendSignal(stopCh)
	}(stopCh)

	select {
	case sig := <-stopCh:
		fmt.Printf("Got %s signal. Aborting...\n", sig)
	case <-ctx.Done():
		fmt.Print("done")
	}
}

func handle() {
	for i := 0; i < 5; i++ {
		fmt.Print("#")
		time.Sleep(time.Millisecond * 100)
	}
	fmt.Println()
}

func sendSignal(stopChan chan os.Signal) {
	fmt.Printf("...")
	time.Sleep(1 * time.Second)
	stopChan <- os.Interrupt
}
