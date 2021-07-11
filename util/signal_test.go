package util

import (
	"fmt"
	"os"
	"testing"
)

func TestSetupSignalHandler(t *testing.T) {
	stopCh := SetupSignalHandler()
	go func() {
		<-stopCh
		fmt.Print("done")
	}()
	InjectSignal(os.Interrupt)
}
