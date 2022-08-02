package pkg

import (
	"github.com/iancoleman/strcase"
	"golang.org/x/sync/syncmap"
)

type OverlayModel struct {
	Controller string
	Name       string
	State      bool

	Red   int
	Green int
	Blue  int
}

type StateMachine struct {
	OverlayModels syncmap.Map
	Callback      OnStateChange
}

type OnStateChange func(OverlayModel)

func (sm *StateMachine) generateSlug(controller, name string) string {
	return strcase.ToSnake(name)
}

func (sm *StateMachine) UpdateState(controller, name string, state bool) {
	slug := sm.generateSlug(controller, name)
	existing, exists := sm.OverlayModels.Load(slug)

	newCopy := OverlayModel{}
	oldCopy := OverlayModel{}

	if !exists {
		oldCopy = OverlayModel{}
		newCopy = OverlayModel{
			Controller: controller,
			Name:       name,
			State:      state,
		}
	} else {
		oldCopy = existing.(OverlayModel)
		newCopy = oldCopy
		newCopy.State = state
	}

	sm.OverlayModels.Store(slug, newCopy)
	if newCopy.State != oldCopy.State {
		sm.Callback(newCopy)
	}
}

func (sm *StateMachine) UpdateColor(controller, name string, red, green, blue int) {
	slug := sm.generateSlug(controller, name)
	existing, exists := sm.OverlayModels.Load(slug)

	newCopy := OverlayModel{}
	oldCopy := OverlayModel{}

	if !exists {
		oldCopy = OverlayModel{}
		newCopy = OverlayModel{
			Controller: controller,
			Name:       name,
			Red:        red,
			Green:      green,
			Blue:       blue,
		}
	} else {
		oldCopy = existing.(OverlayModel)
		newCopy = oldCopy
		newCopy.Red = red
		newCopy.Green = green
		newCopy.Blue = blue
	}

	sm.OverlayModels.Store(slug, newCopy)
	if (newCopy.Red != oldCopy.Red) || (newCopy.Green != oldCopy.Green) || (newCopy.Blue != oldCopy.Blue) {
		sm.Callback(newCopy)
	}
}
