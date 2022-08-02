package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type OverlayModelResponseData struct {
	Name     string `json:"name"`
	IsActive int    `json:"isActive"`
}

type QueryEngine struct {
	Controllers  []string
	Interval     time.Duration
	StateMachine *StateMachine
	Log          *logrus.Entry

	stopChannel chan bool
}

func (qe *QueryEngine) Stop() {
	qe.Log.Debug("Stopping loop...")
	qe.stopChannel <- true
}

func (qe *QueryEngine) Start() error {
	if len(qe.Controllers) == 0 {
		return errors.New("query engine controller list must have at least 1 index")
	}

	if qe.Interval < time.Second {
		return errors.New("query engine interval must be greater than 1 second")
	}

	if qe.StateMachine == nil {
		return errors.New("query engine state machine must not be nil")
	}

	qe.Log.Debugf("Starting loop every %v ...", qe.Interval)
	qe.stopChannel = make(chan bool)
	go qe.runLoop()
	return nil
}

func (qe *QueryEngine) runLoop() {
	for {
		select {
		case <-qe.stopChannel:
			qe.Log.Debug("Exiting query engine loop")
			return
		case <-time.After(qe.Interval):
			qe.queryControllers()
			continue
		}
	}
}

func (qe *QueryEngine) queryControllers() {
	qe.Log.Debug("Querying controllers for information...")

	for _, controller := range qe.Controllers {
		go qe.fetchControllerResponse(controller)
	}
}

func (qe *QueryEngine) fetchControllerResponse(controller string) {
	log := qe.Log.WithField("controller", controller)
	log.Debug("Querying for overlay model states...")

	ctx, cancel := context.WithTimeout(context.Background(), qe.Interval)
	defer cancel()

	url := fmt.Sprintf("http://%s/api/overlays/models", controller)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithError(err).Warn("Could not query the controller state")
		return
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		log.WithError(err).Warn("Could not query the controller state")
		return
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Warnf("Could not read body of controller response")
		return
	}

	defer resp.Body.Close()
	log.Debugf("Fetched models; parsing response...")

	data, err := qe.parseControllerResponse(bytes)
	if err != nil {
		log.WithError(err).Warn("Could not parse controller response")
		return
	}

	qe.updateStateMachineWithControllerData(controller, data)
}

func (qe *QueryEngine) parseControllerResponse(body []byte) ([]OverlayModelResponseData, error) {
	var data []OverlayModelResponseData

	if err := json.Unmarshal(body, &data); err != nil {
		return data, err
	}

	return data, nil
}

func (qe *QueryEngine) updateStateMachineWithControllerData(controller string, data []OverlayModelResponseData) {
	for _, overlayModel := range data {
		state := false
		if overlayModel.IsActive > 0 {
			state = true
		}

		qe.StateMachine.UpdateState(controller, overlayModel.Name, state)
	}
}
