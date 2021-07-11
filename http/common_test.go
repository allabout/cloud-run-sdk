package http

import (
	"bytes"
	"os"
	"testing"

	pkgzerolog "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	// logger io.Writer to buffer to disable display log after terminate server
	buf := &bytes.Buffer{}
	log.Logger = pkgzerolog.New(buf).With().Timestamp().Logger()

	if err := os.Setenv("K_CONFIGURATION", "true"); err != nil {
		log.Fatal().Msgf("%v", err)
	}
	m.Run()
}
