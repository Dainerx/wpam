package main

import (
	"os"

	"github.com/Dainerx/wpam/cmd"
	"github.com/Dainerx/wpam/pkg/logger"
	log "github.com/sirupsen/logrus"
)

const LogFile = "log.log"

func init() {
	// Init logger
	//File Path for logging
	//var logFilePath string
	//File for logging
	f, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(f)

	log.SetLevel(log.InfoLevel) // Only log the info severity or above.

	logger.Logger = *log.WithFields(log.Fields{})
}

func main() {
	err := cmd.Execute()
	if err != nil {
		logger.Logger.Fatalf("%v", err)
	}
}
