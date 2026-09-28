package color

import (
	"fmt"
	"sync"

	"github.com/wyw14/solo-038/internal/model"
)

type State struct {
	mu          sync.RWMutex
	baseline    model.ColorBaseline
	established bool
	readings    []model.ColorReading
}

func NewState() *State {
	return &State{}
}

func (state *State) Establish(baseline model.ColorBaseline) error {
	if err := baseline.Validate(); err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	state.baseline = baseline
	state.established = true
	state.readings = nil
	return nil
}

func (state *State) Add(reading model.ColorReading) error {
	if err := reading.Validate(); err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if !state.established {
		return fmt.Errorf("color baseline is not established")
	}
	if !state.baseline.Accepts(reading) {
		return fmt.Errorf("color reading does not belong to active baseline; re-establish baseline after ink batch or plate change")
	}
	if len(state.readings) > 0 && reading.WebMeter < state.readings[len(state.readings)-1].WebMeter {
		return fmt.Errorf("color reading meter moved backwards")
	}
	state.readings = append(state.readings, reading)
	return nil
}

func (state *State) Snapshot() (model.ColorBaseline, []model.ColorReading, bool) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	readings := append([]model.ColorReading(nil), state.readings...)
	return state.baseline, readings, state.established
}
