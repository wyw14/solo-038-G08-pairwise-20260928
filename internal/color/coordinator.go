package color

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/wyw14/solo-038/internal/journal"
	"github.com/wyw14/solo-038/internal/model"
)

type ExceededError struct {
	Segments []model.ColorSegment `json:"segments"`
}

func (err *ExceededError) Error() string {
	return fmt.Sprintf("color deviation exceeds tolerance in %d segment(s)", len(err.Segments))
}

type Coordinator struct {
	mu        sync.Mutex
	state     *State
	journal   *journal.Store
	lastError error
}

func NewCoordinator(state *State, store *journal.Store) *Coordinator {
	return &Coordinator{state: state, journal: store}
}

func (coordinator *Coordinator) EstablishBaseline(job model.Job, operation model.OperationID, target model.ColorTarget, inkBatch string, now time.Time) (model.ColorBaseline, error) {
	baseline := model.ColorBaseline{JobGeneration: job.Generation, PlateGeneration: job.PlateGeneration, InkBatch: inkBatch, Target: target, EstablishedAt: now.UTC()}
	if err := coordinator.state.Establish(baseline); err != nil {
		return model.ColorBaseline{}, err
	}
	if _, err := journal.PersistColorBaseline(coordinator.journal, job, operation, baseline, now); err != nil {
		coordinator.setLastError(err)
		return model.ColorBaseline{}, err
	}
	return baseline, nil
}

func (coordinator *Coordinator) AcceptReading(reading model.ColorReading) (model.ColorDeviation, error) {
	if err := coordinator.state.Add(reading); err != nil {
		return model.ColorDeviation{}, err
	}
	baseline, _, _ := coordinator.state.Snapshot()
	return Deviation(baseline.Target, reading), nil
}

func (coordinator *Coordinator) Baseline() (model.ColorBaseline, bool) {
	baseline, _, established := coordinator.state.Snapshot()
	return baseline, established
}

func (coordinator *Coordinator) Evaluate() ([]model.ColorDeviation, []model.ColorSegment, error) {
	baseline, readings, established := coordinator.state.Snapshot()
	if !established {
		return nil, nil, fmt.Errorf("color baseline is not established")
	}
	deviations := make([]model.ColorDeviation, 0, len(readings))
	for _, reading := range readings {
		deviations = append(deviations, Deviation(baseline.Target, reading))
	}
	return deviations, MergeSegments(deviations), nil
}

func (coordinator *Coordinator) Release(job model.Job, operation model.OperationID, now time.Time) (journal.ColorReleaseRecord, error) {
	baseline, readings, established := coordinator.state.Snapshot()
	if !established {
		return journal.ColorReleaseRecord{}, fmt.Errorf("color baseline is not established")
	}
	if baseline.JobGeneration != job.Generation || baseline.PlateGeneration != job.PlateGeneration {
		return journal.ColorReleaseRecord{}, fmt.Errorf("color baseline is stale; re-establish baseline after ink batch or plate change")
	}
	if len(readings) < 2 {
		return journal.ColorReleaseRecord{}, fmt.Errorf("color readings are incomplete")
	}
	deviations := make([]model.ColorDeviation, 0, len(readings))
	for _, reading := range readings {
		deviations = append(deviations, Deviation(baseline.Target, reading))
	}
	segments := MergeSegments(deviations)
	if len(segments) > 0 {
		return journal.ColorReleaseRecord{}, &ExceededError{Segments: segments}
	}
	interval := model.Interval{JobID: job.ID, JobGeneration: job.Generation, RollGeneration: job.RollGeneration, StartMeter: readings[0].WebMeter, EndMeter: readings[len(readings)-1].WebMeter, OpenedAt: readings[0].MeasuredAt, ClosedAt: now.UTC()}
	record := journal.ColorReleaseRecord{Baseline: baseline, Interval: interval, Readings: readings, Deviations: deviations, Released: true}
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	if _, err := journal.PersistColorRelease(coordinator.journal, job, operation, record, now); err != nil {
		coordinator.lastError = err
		return journal.ColorReleaseRecord{}, err
	}
	return record, nil
}

func (coordinator *Coordinator) LastError() error {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	return coordinator.lastError
}

func (coordinator *Coordinator) setLastError(err error) {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	coordinator.lastError = err
}

func Deviation(target model.ColorTarget, reading model.ColorReading) model.ColorDeviation {
	deltaE := target.Formula.DeltaE(target.Lab, reading.Lab)
	return model.ColorDeviation{ReadingID: reading.ID, WebMeter: reading.WebMeter, DeltaE: deltaE, Limit: target.Tolerance, Exceeds: deltaE > target.Tolerance}
}

func MergeSegments(deviations []model.ColorDeviation) []model.ColorSegment {
	segments := []model.ColorSegment{}
	open := -1
	for _, deviation := range deviations {
		if !deviation.Exceeds {
			open = -1
			continue
		}
		if open < 0 {
			segments = append(segments, model.ColorSegment{StartMeter: deviation.WebMeter, EndMeter: deviation.WebMeter})
			open = len(segments) - 1
		}
		segment := &segments[open]
		segment.EndMeter = deviation.WebMeter
		segment.Readings++
		segment.MaxDeltaE = math.Max(segment.MaxDeltaE, deviation.DeltaE)
	}
	return segments
}
