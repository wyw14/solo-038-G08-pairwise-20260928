package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wyw14/solo-038/internal/color"
	"github.com/wyw14/solo-038/internal/journal"
	"github.com/wyw14/solo-038/internal/model"
)

func (runtime *Runtime) EstablishColorBaseline(target model.Lab, tolerance float64, formula string) (model.Baseline, error) {
	current, ok := runtime.jobs.Current()
	if !ok {
		return model.Baseline{}, fmt.Errorf("there is no active job")
	}
	return runtime.colorControl.EstablishBaseline(current, runtime.ink.Batch(), formula, target, tolerance, time.Now())
}

func (runtime *Runtime) RecordColorMeasurements(samples []color.Sample) ([]model.Measurement, error) {
	current, ok := runtime.jobs.Current()
	if !ok {
		return nil, fmt.Errorf("there is no active job")
	}
	measurements, err := runtime.colorControl.Record(current, runtime.ink.Batch(), samples, time.Now())
	if err != nil {
		return nil, err
	}
	outside := 0
	for _, measurement := range measurements {
		if !measurement.Within {
			outside++
		}
	}
	if outside > 0 {
		runtime.RecordIncident("warning", "color", fmt.Sprintf("%d color measurements out of tolerance", outside))
	}
	return measurements, nil
}

type ColorStatus struct {
	Baseline      *model.Baseline `json:"baseline"`
	BaselineStale bool            `json:"baseline_stale"`
	InkBatch      string          `json:"ink_batch"`
	Measurements  int             `json:"measurements"`
	Segments      []model.Segment `json:"out_of_tolerance"`
}

func (runtime *Runtime) ColorStatus() ColorStatus {
	status := ColorStatus{InkBatch: runtime.ink.Batch(), Segments: runtime.colorControl.Segments()}
	baseline, ok := runtime.colorControl.Baseline()
	if !ok {
		return status
	}
	status.Baseline = &baseline
	status.Measurements = len(runtime.colorControl.Measurements())
	if current, ok := runtime.jobs.Current(); ok {
		status.BaselineStale = !baseline.Matches(current.Generation, current.PlateGeneration, runtime.ink.Batch())
	}
	return status
}

func (runtime *Runtime) ReleaseColor(ctx context.Context) (journal.ColorReleaseRecord, error) {
	select {
	case <-ctx.Done():
		return journal.ColorReleaseRecord{}, ctx.Err()
	default:
	}
	current, ok := runtime.jobs.Current()
	if !ok {
		return journal.ColorReleaseRecord{}, fmt.Errorf("there is no active job")
	}
	record, err := runtime.colorControl.Release(current, model.NewOperationID(), runtime.ink.Batch(), time.Now())
	if err != nil && record.BaselineID != "" && !record.Released {
		runtime.RecordIncident("error", "color", fmt.Sprintf("color release rejected: %d out-of-tolerance segments", len(record.Segments)))
	}
	return record, err
}

func (runtime *Runtime) ChangeInkBatch(batch string) error {
	if strings.TrimSpace(batch) == "" {
		return fmt.Errorf("ink batch is required")
	}
	runtime.ink.SetBatch(batch)
	return nil
}
