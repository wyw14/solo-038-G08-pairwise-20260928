package journal

import (
	"fmt"
	"time"

	"github.com/wyw14/solo-038/internal/model"
)

type IntervalRecord struct {
	Interval model.Interval `json:"interval"`
	Proofs   []model.Proof  `json:"proofs"`
	Released bool           `json:"released"`
}

func PersistInterval(store *Store, job model.Job, operation model.OperationID, record IntervalRecord, now time.Time) (uint64, error) {
	if err := record.Interval.Validate(); err != nil {
		return 0, err
	}
	event, err := model.NewEvent(operation, job, "interval_committed", record, now)
	if err != nil {
		return 0, err
	}
	sequence, err := store.Append(event)
	if err != nil {
		return sequence, err
	}
	if !store.State().IsDurable(sequence) {
		return sequence, fmt.Errorf("interval sequence %d is not durable", sequence)
	}
	return sequence, nil
}

func PersistTransition(store *Store, job model.Job, operation model.OperationID, from, to model.JobState, now time.Time) (uint64, error) {
	payload := map[string]any{"from": from, "to": to, "generation": job.Generation}
	event, err := model.NewEvent(operation, job, "job_transitioned", payload, now)
	if err != nil {
		return 0, err
	}
	return store.Append(event)
}

func PersistFence(store *Store, job model.Job, operation model.OperationID, owner string, now time.Time) (uint64, error) {
	payload := map[string]any{"owner": owner, "roll_generation": job.RollGeneration}
	event, err := model.NewEvent(operation, job, "web_fence_restored", payload, now)
	if err != nil {
		return 0, err
	}
	return store.Append(event)
}

type ColorSamplesRecord struct {
	Measurements []model.Measurement `json:"measurements"`
}

type ColorReleaseRecord struct {
	BaselineID      string          `json:"baseline_id"`
	JobGeneration   uint64          `json:"job_generation"`
	PlateGeneration uint64          `json:"plate_generation"`
	InkBatch        string          `json:"ink_batch"`
	StartMeter      float64         `json:"start_meter"`
	EndMeter        float64         `json:"end_meter"`
	Released        bool            `json:"released"`
	Segments        []model.Segment `json:"segments"`
	MeasurementIDs  []string        `json:"measurement_ids"`
	DecidedAt       time.Time       `json:"decided_at"`
}

func PersistColorBaseline(store *Store, job model.Job, operation model.OperationID, baseline model.Baseline, now time.Time) (uint64, error) {
	if err := baseline.Validate(); err != nil {
		return 0, err
	}
	event, err := model.NewEvent(operation, job, "color_baseline_established", baseline, now)
	if err != nil {
		return 0, err
	}
	return store.Append(event)
}

func PersistColorSamples(store *Store, job model.Job, operation model.OperationID, measurements []model.Measurement, now time.Time) (uint64, error) {
	if len(measurements) == 0 {
		return 0, fmt.Errorf("color samples are required")
	}
	for _, measurement := range measurements {
		if err := measurement.Validate(); err != nil {
			return 0, err
		}
	}
	event, err := model.NewEvent(operation, job, "color_samples_recorded", ColorSamplesRecord{Measurements: measurements}, now)
	if err != nil {
		return 0, err
	}
	return store.Append(event)
}

func PersistColorRelease(store *Store, job model.Job, operation model.OperationID, record ColorReleaseRecord, now time.Time) (uint64, error) {
	if record.BaselineID == "" || len(record.MeasurementIDs) == 0 {
		return 0, fmt.Errorf("color release record is incomplete")
	}
	event, err := model.NewEvent(operation, job, "color_release_decided", record, now)
	if err != nil {
		return 0, err
	}
	sequence, err := store.Append(event)
	if err != nil {
		return sequence, err
	}
	if !store.State().IsDurable(sequence) {
		return sequence, fmt.Errorf("color release sequence %d is not durable", sequence)
	}
	return sequence, nil
}
