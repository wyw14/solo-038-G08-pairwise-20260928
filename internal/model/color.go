package model

import (
	"errors"
	"fmt"
	"time"
)

type Lab struct {
	L float64 `json:"l"`
	A float64 `json:"a"`
	B float64 `json:"b"`
}

func (lab Lab) Validate() error {
	if lab.L < 0 || lab.L > 100 {
		return fmt.Errorf("lab lightness %v is outside 0..100", lab.L)
	}
	if lab.A < -128 || lab.A > 127 || lab.B < -128 || lab.B > 127 {
		return fmt.Errorf("lab chroma coordinates %v/%v are outside -128..127", lab.A, lab.B)
	}
	return nil
}

type Baseline struct {
	ID              string    `json:"id"`
	Target          Lab       `json:"target"`
	Tolerance       float64   `json:"tolerance"`
	Formula         string    `json:"formula"`
	JobGeneration   uint64    `json:"job_generation"`
	PlateGeneration uint64    `json:"plate_generation"`
	InkBatch        string    `json:"ink_batch"`
	EstablishedAt   time.Time `json:"established_at"`
}

func (baseline Baseline) Validate() error {
	if baseline.ID == "" {
		return errors.New("baseline id is required")
	}
	if err := baseline.Target.Validate(); err != nil {
		return err
	}
	if baseline.Tolerance <= 0 {
		return errors.New("baseline tolerance is required")
	}
	if baseline.Formula == "" {
		return errors.New("baseline formula is required")
	}
	if baseline.JobGeneration == 0 || baseline.PlateGeneration == 0 {
		return errors.New("baseline generations are required")
	}
	if baseline.InkBatch == "" {
		return errors.New("baseline ink batch is required")
	}
	if baseline.EstablishedAt.IsZero() {
		return errors.New("baseline timestamp is required")
	}
	return nil
}

func (baseline Baseline) Matches(jobGeneration, plateGeneration uint64, inkBatch string) bool {
	return baseline.JobGeneration == jobGeneration && baseline.PlateGeneration == plateGeneration && baseline.InkBatch == inkBatch
}

type Measurement struct {
	ID              string    `json:"id"`
	BaselineID      string    `json:"baseline_id"`
	JobGeneration   uint64    `json:"job_generation"`
	PlateGeneration uint64    `json:"plate_generation"`
	WebMeter        float64   `json:"web_meter"`
	Measured        Lab       `json:"measured"`
	DeltaE          float64   `json:"delta_e"`
	Within          bool      `json:"within"`
	ObservedAt      time.Time `json:"observed_at"`
}

func (measurement Measurement) Validate() error {
	if measurement.ID == "" || measurement.BaselineID == "" {
		return errors.New("measurement identity is incomplete")
	}
	if measurement.JobGeneration == 0 || measurement.PlateGeneration == 0 {
		return errors.New("measurement generations are required")
	}
	if measurement.WebMeter < 0 {
		return errors.New("measurement web meter is invalid")
	}
	if err := measurement.Measured.Validate(); err != nil {
		return err
	}
	if measurement.DeltaE < 0 {
		return errors.New("measurement delta e is invalid")
	}
	if measurement.ObservedAt.IsZero() {
		return errors.New("measurement timestamp is required")
	}
	return nil
}

type Segment struct {
	StartMeter     float64  `json:"start_meter"`
	EndMeter       float64  `json:"end_meter"`
	MaxDeltaE      float64  `json:"max_delta_e"`
	Samples        int      `json:"samples"`
	MeasurementIDs []string `json:"measurement_ids"`
}
