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

type ColorReleaseRecord struct {
	Baseline   model.ColorBaseline    `json:"baseline"`
	Interval   model.Interval         `json:"interval"`
	Readings   []model.ColorReading   `json:"readings"`
	Deviations []model.ColorDeviation `json:"deviations"`
	Released   bool                   `json:"released"`
}

func (record ColorReleaseRecord) Validate() error {
	if err := record.Baseline.Validate(); err != nil {
		return err
	}
	if err := record.Interval.Validate(); err != nil {
		return err
	}
	if len(record.Readings) == 0 || len(record.Readings) != len(record.Deviations) {
		return fmt.Errorf("color release evidence is incomplete")
	}
	for _, reading := range record.Readings {
		if err := reading.Validate(); err != nil {
			return err
		}
	}
	return nil
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

func PersistColorRelease(store *Store, job model.Job, operation model.OperationID, record ColorReleaseRecord, now time.Time) (uint64, error) {
	if err := record.Validate(); err != nil {
		return 0, err
	}
	event, err := model.NewEvent(operation, job, "color_released", record, now)
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

func PersistColorBaseline(store *Store, job model.Job, operation model.OperationID, baseline model.ColorBaseline, now time.Time) (uint64, error) {
	if err := baseline.Validate(); err != nil {
		return 0, err
	}
	event, err := model.NewEvent(operation, job, "color_baseline_established", baseline, now)
	if err != nil {
		return 0, err
	}
	return store.Append(event)
}
