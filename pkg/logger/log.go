package logger

import (
	log "github.com/sirupsen/logrus"
)

var Logger log.Entry

func init() {
	if Logger.Logger == nil {
		log.SetLevel(log.ErrorLevel) // Only log the info severity or above.
		Logger = *log.WithFields(log.Fields{})
	}
}
