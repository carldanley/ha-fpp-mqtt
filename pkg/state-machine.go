package pkg

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"golang.org/x/sync/syncmap"
)

type OverlayModel struct {
	Name     string
	State    bool
	RGBColor []int
}

type StateMachine struct {
	OverlayModels syncmap.Map
	Callback      OnStateChange
}

type OnStateChange func(string, OverlayModel)

func (sm *StateMachine) GenerateSlug(controller, name string) string {
	return fmt.Sprintf("%s-%s", strcase.ToSnake(controller), strcase.ToSnake(name))
}

func (sm *StateMachine) createNewCopy(slug, name string, state bool, rgbColor []int) (newCopy, oldCopy OverlayModel) {
	existing, exists := sm.OverlayModels.Load(slug)

	if !exists {
		newCopy = OverlayModel{
			Name:  name,
			State: state,
		}

		if len(rgbColor) == ExpectedColorCount {
			newCopy.RGBColor = rgbColor
		}
	} else {
		oldCopy = existing.(OverlayModel)
		newCopy = oldCopy

		if name != "" {
			newCopy.Name = name
		}

		newCopy.State = state

		if len(rgbColor) == ExpectedColorCount {
			newCopy.RGBColor = rgbColor
		}
	}

	return newCopy, oldCopy
}

func (sm *StateMachine) UpdateBySlug(slug, name string, state bool, rgbColor []int) {
	newCopy, oldCopy := sm.createNewCopy(slug, name, state, rgbColor)

	sm.OverlayModels.Store(slug, newCopy)

	statesChanged := newCopy.State != oldCopy.State
	colorsChanged := false

	if len(newCopy.RGBColor) != len(oldCopy.RGBColor) {
		colorsChanged = true
	} else if len(newCopy.RGBColor) == ExpectedColorCount {
		if newCopy.RGBColor[0] != oldCopy.RGBColor[0] {
			colorsChanged = true
		} else if newCopy.RGBColor[1] != oldCopy.RGBColor[1] {
			colorsChanged = true
		} else if newCopy.RGBColor[2] != oldCopy.RGBColor[2] {
			colorsChanged = true
		}
	}

	if statesChanged || colorsChanged {
		sm.Callback(slug, newCopy)
	}
}
