package color

import (
	"sync"

	"github.com/wyw14/solo-038/internal/model"
)

const measurementLimit = 500

type State struct {
	mu           sync.RWMutex
	baseline     model.Baseline
	hasBaseline  bool
	measurements []model.Measurement
}

func NewState() *State {
	return &State{}
}

func (state *State) PutBaseline(baseline model.Baseline) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.baseline = baseline
	state.hasBaseline = true
}

func (state *State) Baseline() (model.Baseline, bool) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.baseline, state.hasBaseline
}

func (state *State) Add(measurement model.Measurement) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.measurements = append(state.measurements, measurement)
	if len(state.measurements) > measurementLimit {
		state.measurements = state.measurements[len(state.measurements)-measurementLimit:]
	}
}

func (state *State) Measurements() []model.Measurement {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return append([]model.Measurement(nil), state.measurements...)
}

func (state *State) BaselineMeasurements(baselineID string) []model.Measurement {
	state.mu.RLock()
	defer state.mu.RUnlock()
	values := []model.Measurement{}
	for _, measurement := range state.measurements {
		if measurement.BaselineID == baselineID {
			values = append(values, measurement)
		}
	}
	return values
}
