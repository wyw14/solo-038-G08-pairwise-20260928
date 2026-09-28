package color

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/solo-038/internal/journal"
	"github.com/wyw14/solo-038/internal/model"
)

type Sample struct {
	WebMeter float64   `json:"web_meter"`
	Measured model.Lab `json:"measured"`
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

func (coordinator *Coordinator) EstablishBaseline(job model.Job, inkBatch, formula string, target model.Lab, tolerance float64, now time.Time) (model.Baseline, error) {
	if formula == "" {
		formula = FormulaCIEDE2000
	}
	if _, err := DeltaE(formula, target, target); err != nil {
		return model.Baseline{}, err
	}
	baseline := model.Baseline{ID: string(model.NewEventID()), Target: target, Tolerance: tolerance, Formula: formula, JobGeneration: job.Generation, PlateGeneration: job.PlateGeneration, InkBatch: inkBatch, EstablishedAt: now.UTC()}
	if err := baseline.Validate(); err != nil {
		return model.Baseline{}, err
	}
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	coordinator.state.PutBaseline(baseline)
	if _, err := journal.PersistColorBaseline(coordinator.journal, job, model.NewOperationID(), baseline, now); err != nil {
		coordinator.lastError = err
		return model.Baseline{}, err
	}
	return baseline, nil
}

func (coordinator *Coordinator) Record(job model.Job, inkBatch string, samples []Sample, now time.Time) ([]model.Measurement, error) {
	if len(samples) == 0 {
		return nil, fmt.Errorf("color samples are required")
	}
	baseline, ok := coordinator.state.Baseline()
	if !ok {
		return nil, fmt.Errorf("color baseline is not established")
	}
	if !baseline.Matches(job.Generation, job.PlateGeneration, inkBatch) {
		return nil, fmt.Errorf("color baseline is stale; re-establish after ink batch or plate change")
	}
	measurements := make([]model.Measurement, 0, len(samples))
	for _, sample := range samples {
		deltaE, err := DeltaE(baseline.Formula, baseline.Target, sample.Measured)
		if err != nil {
			return nil, err
		}
		measurement := model.Measurement{ID: string(model.NewEventID()), BaselineID: baseline.ID, JobGeneration: job.Generation, PlateGeneration: job.PlateGeneration, WebMeter: sample.WebMeter, Measured: sample.Measured, DeltaE: deltaE, Within: deltaE <= baseline.Tolerance, ObservedAt: now.UTC()}
		if err := measurement.Validate(); err != nil {
			return nil, err
		}
		measurements = append(measurements, measurement)
	}
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	for _, measurement := range measurements {
		coordinator.state.Add(measurement)
	}
	if _, err := journal.PersistColorSamples(coordinator.journal, job, model.NewOperationID(), measurements, now); err != nil {
		coordinator.lastError = err
		return nil, err
	}
	return measurements, nil
}

func (coordinator *Coordinator) Baseline() (model.Baseline, bool) {
	return coordinator.state.Baseline()
}

func (coordinator *Coordinator) Measurements() []model.Measurement {
	baseline, ok := coordinator.state.Baseline()
	if !ok {
		return nil
	}
	return coordinator.state.BaselineMeasurements(baseline.ID)
}

func (coordinator *Coordinator) Segments() []model.Segment {
	return BuildSegments(coordinator.Measurements())
}

func (coordinator *Coordinator) Release(job model.Job, operation model.OperationID, inkBatch string, now time.Time) (journal.ColorReleaseRecord, error) {
	baseline, ok := coordinator.state.Baseline()
	if !ok {
		return journal.ColorReleaseRecord{}, fmt.Errorf("color baseline is not established")
	}
	if !baseline.Matches(job.Generation, job.PlateGeneration, inkBatch) {
		return journal.ColorReleaseRecord{}, fmt.Errorf("color baseline is stale; re-establish after ink batch or plate change")
	}
	measurements := coordinator.state.BaselineMeasurements(baseline.ID)
	if len(measurements) == 0 {
		return journal.ColorReleaseRecord{}, fmt.Errorf("no color measurements for active baseline")
	}
	segments := BuildSegments(measurements)
	startMeter, endMeter := meterSpan(measurements)
	record := journal.ColorReleaseRecord{BaselineID: baseline.ID, JobGeneration: job.Generation, PlateGeneration: job.PlateGeneration, InkBatch: inkBatch, StartMeter: startMeter, EndMeter: endMeter, Released: len(segments) == 0, Segments: segments, MeasurementIDs: measurementIDs(measurements), DecidedAt: now.UTC()}
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	if _, err := journal.PersistColorRelease(coordinator.journal, job, operation, record, now); err != nil {
		coordinator.lastError = err
		return record, err
	}
	if !record.Released {
		return record, fmt.Errorf("color release rejected: %d out-of-tolerance segments", len(segments))
	}
	return record, nil
}

func (coordinator *Coordinator) LastError() error {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	return coordinator.lastError
}
