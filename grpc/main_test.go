package grpc

import (
	"bytes"
	"os"
	"testing"

	pkgzerolog "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logBuffer = &bytes.Buffer{}

func TestMain(m *testing.M) {
	// logger io.Writer to buffer to disable display log after terminate server
	log.Logger = pkgzerolog.New(logBuffer).With().Timestamp().Logger()

	if err := os.Setenv("K_CONFIGURATION", "true"); err != nil {
		log.Fatal().Msgf("%v", err)
	}
	m.Run()
}
