package service

import (
	"fmt"
	"time"

	"github.com/wyw14/solo-038/internal/color"
	"github.com/wyw14/solo-038/internal/journal"
	"github.com/wyw14/solo-038/internal/model"
)

type ColorStatus struct {
	Baseline    model.ColorBaseline    `json:"baseline"`
	Established bool                   `json:"established"`
	Readings    int                    `json:"readings"`
	Deviations  []model.ColorDeviation `json:"deviations"`
	Segments    []model.ColorSegment   `json:"segments"`
}

func (runtime *Runtime) EstablishColorBaseline(target model.ColorTarget, inkBatch string) (model.ColorBaseline, error) {
	current, ok := runtime.jobs.Current()
	if !ok {
		return model.ColorBaseline{}, fmt.Errorf("there is no active job")
	}
	return runtime.colorControl.EstablishBaseline(current, model.NewOperationID(), target, inkBatch, time.Now())
}

func (runtime *Runtime) AcceptColorReading(inkBatch string, webMeter float64, lab model.LabColor, measuredAt time.Time) (model.ColorDeviation, error) {
	current, ok := runtime.jobs.Current()
	if !ok {
		return model.ColorDeviation{}, fmt.Errorf("there is no active job")
	}
	if measuredAt.IsZero() {
		measuredAt = time.Now()
	}
	reading := model.ColorReading{ID: string(model.NewOperationID()), JobGeneration: current.Generation, PlateGeneration: current.PlateGeneration, InkBatch: inkBatch, WebMeter: webMeter, Lab: lab, MeasuredAt: measuredAt.UTC()}
	return runtime.colorControl.AcceptReading(reading)
}

func (runtime *Runtime) ColorStatus() ColorStatus {
	baseline, established := runtime.colorControl.Baseline()
	status := ColorStatus{Baseline: baseline, Established: established, Deviations: []model.ColorDeviation{}, Segments: []model.ColorSegment{}}
	deviations, segments, err := runtime.colorControl.Evaluate()
	if err != nil {
		return status
	}
	status.Readings = len(deviations)
	status.Deviations = deviations
	status.Segments = segments
	return status
}

func (runtime *Runtime) ReleaseColor() (journal.ColorReleaseRecord, error) {
	current, ok := runtime.jobs.Current()
	if !ok {
		return journal.ColorReleaseRecord{}, fmt.Errorf("there is no active job")
	}
	return runtime.colorControl.Release(current, model.NewOperationID(), time.Now())
}

func (runtime *Runtime) ColorReleases() ([]journal.ColorReleaseRecord, error) {
	releases, err := color.Releases(runtime.journal)
	if err != nil {
		return nil, err
	}
	if releases == nil {
		releases = []journal.ColorReleaseRecord{}
	}
	return releases, nil
}
