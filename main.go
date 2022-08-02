package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var signalChannel chan os.Signal
var log *logrus.Logger

func init() {
	// create our channel for signal interrupts
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT)

	// setup the logging level & controller list (csv)
	logLevel := flag.Int("v", 0, "the level of verbosity for logging")
	flag.Parse()

	// create a new logger
	log = logrus.New()
	log.SetOutput(os.Stdout)
	// set the log level
	if *logLevel == 0 {
		log.SetLevel(logrus.ErrorLevel)
	} else if *logLevel == 1 {
		log.SetLevel(logrus.WarnLevel)
	} else if *logLevel == 2 {
		log.SetLevel(logrus.InfoLevel)
	} else if *logLevel >= 3 {
		log.SetLevel(logrus.DebugLevel)
	}

	// figure out where to load the .env file from
	envFileLocation := ".env"
	if location, exists := os.LookupEnv("ENV_FILE"); exists {
		envFileLocation = location
	}

	// attempt to load our .env file
	log.WithField("location", envFileLocation).Debug("Loading environment file from location")
	if err := godotenv.Load(envFileLocation); err != nil {
		log.WithError(err).Warn("Could not load environment file from location")
	}
}

func main() {
	// wait for an interrupt signal
	<-signalChannel
}
