package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"git.r1p.io/alfred/ha-fpp-mqtt/pkg"
	"github.com/davecgh/go-spew/spew"
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

func parseControllerList() []string {
	controllers := []string{}
	parts := strings.Split(os.Getenv("CONTROLLER_LIST"), ",")

	for _, controller := range parts {
		controller = strings.ReplaceAll(controller, " ", "")
		controllers = append(controllers, controller)
	}

	return controllers
}

func parseQueryInterval() time.Duration {
	secondsString, ok := os.LookupEnv("QUERY_INTERVAL_SECONDS")
	if !ok {
		return time.Second
	}

	seconds, err := strconv.Atoi(secondsString)
	if err != nil {
		return time.Second
	}

	return time.Second * time.Duration(seconds)
}

func publishOverlayModelStatus(model pkg.OverlayModel) {
	spew.Dump(model)
}

func main() {
	// grab the list of controllers we need to monitor state for
	controllers := parseControllerList()
	if len(controllers) == 0 {
		log.Fatal("invalid controller list")
	}

	// create a new state machine
	stateMachine := pkg.StateMachine{
		Callback: publishOverlayModelStatus,
	}

	// create a new query engine
	queryEngine := pkg.QueryEngine{
		Controllers:  controllers,
		Interval:     parseQueryInterval(),
		StateMachine: &stateMachine,
		Log:          log.WithField("component", "queryEngine"),
	}

	// startup the query engine
	if err := queryEngine.Start(); err != nil {
		log.WithError(err).Fatal("Could not start query engine")
	}

	// wait for an interrupt signal
	<-signalChannel
	queryEngine.Stop()
}
