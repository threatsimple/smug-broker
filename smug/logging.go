// wrapper around our logging so we can setup certain attributes easily

package smug

import (
	"os"

	log "github.com/sirupsen/logrus"
)

type Logger struct {
	log.Entry
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)
}

func SetupLogging(loglevel string) {
	if lev, err := log.ParseLevel(loglevel); err == nil {
		log.SetLevel(lev)
	} else {
		log.Panic("invalid loglevel")
	}
}

func NewLogger(context string) *Logger {
	return &Logger{*log.WithFields(log.Fields{"ctx": context})}
}
