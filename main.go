package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/carldanley/ha-fpp-mqtt/pkg"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const LogLevelError = 0
const LogLevelWarn = 1
const LogLevelInfo = 2
const LogLevelDebug = 3

func main() {
	var signalChannel chan os.Signal

	var log *logrus.Logger

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
	if *logLevel == LogLevelError {
		log.SetLevel(logrus.ErrorLevel)
	} else if *logLevel == LogLevelWarn {
		log.SetLevel(logrus.WarnLevel)
	} else if *logLevel == LogLevelInfo {
		log.SetLevel(logrus.InfoLevel)
	} else if *logLevel >= LogLevelDebug {
		log.SetLevel(logrus.DebugLevel)
	}

	// figure out where to load the .env file from
	envFileLocation := ".env"
	if location, exists := os.LookupEnv("ENV_FILE"); exists {
		envFileLocation = location
	}

	// attempt to load our .env file
	log.WithField("location", envFileLocation).Debug("Loading environment file from location")

	err := godotenv.Load(envFileLocation)
	if err != nil {
		log.WithError(err).Warn("Could not load environment file from location")
	}

	// get the config first
	cfg := pkg.GetConfig()

	// setup an mqtt client
	mqttClient := pkg.NewMQTTClient(
		log, cfg.MQTTHost, cfg.MQTTPort, cfg.MQTTClientID,
		cfg.MQTTUsername, cfg.MQTTPassword, cfg.MQTTTopicPrefix,
	)

	defer mqttClient.Stop()

	// create a new state machine
	stateMachine := pkg.StateMachine{
		Callback: mqttClient.PublishOverlayModelStatus,
	}

	// create a new query engine
	queryEngine := pkg.QueryEngine{
		Controllers:  cfg.Controllers,
		Interval:     cfg.QueryInterval,
		StateMachine: &stateMachine,
		Log:          log.WithField("component", "queryEngine"),
	}

	defer queryEngine.Stop()

	// startup the query engine
	if err := queryEngine.Start(); err != nil {
		log.WithError(err).Fatal("Could not start query engine")
	}

	// startup the mqtt client
	if err := mqttClient.Start(&stateMachine); err != nil {
		log.WithError(err).Fatal("Could not start mqtt client")
	}

	// wait for an interrupt signal
	<-signalChannel
}
