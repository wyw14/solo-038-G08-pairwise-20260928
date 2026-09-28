package model

import (
	"errors"
	"math"
	"time"
)

type LabColor struct {
	L float64 `json:"l"`
	A float64 `json:"a"`
	B float64 `json:"b"`
}

func (color LabColor) Validate() error {
	if color.L < 0 || color.L > 100 || color.A < -128 || color.A > 127 || color.B < -128 || color.B > 127 {
		return errors.New("lab color is outside valid range")
	}
	return nil
}

type ColorFormula string

const (
	ColorFormulaCIE76   ColorFormula = "cie76"
	ColorFormulaCIE2000 ColorFormula = "cie2000"
)

func (formula ColorFormula) Valid() bool {
	return formula == ColorFormulaCIE76 || formula == ColorFormulaCIE2000
}

func (formula ColorFormula) DeltaE(target, measured LabColor) float64 {
	if formula == ColorFormulaCIE76 {
		return DeltaE76(target, measured)
	}
	return DeltaE2000(target, measured)
}

func DeltaE76(target, measured LabColor) float64 {
	return math.Sqrt(square(target.L-measured.L) + square(target.A-measured.A) + square(target.B-measured.B))
}

func DeltaE2000(target, measured LabColor) float64 {
	c1 := math.Hypot(target.A, target.B)
	c2 := math.Hypot(measured.A, measured.B)
	cMean := (c1 + c2) / 2
	cMean7 := math.Pow(cMean, 7)
	g := 0.5 * (1 - math.Sqrt(cMean7/(cMean7+math.Pow(25, 7))))
	a1 := (1 + g) * target.A
	a2 := (1 + g) * measured.A
	c1p := math.Hypot(a1, target.B)
	c2p := math.Hypot(a2, measured.B)
	h1p := hueAngle(target.B, a1)
	h2p := hueAngle(measured.B, a2)
	deltaLp := measured.L - target.L
	deltaCp := c2p - c1p
	var deltahp float64
	if c1p*c2p != 0 {
		deltahp = h2p - h1p
		if deltahp > 180 {
			deltahp -= 360
		} else if deltahp < -180 {
			deltahp += 360
		}
	}
	deltaHp := 2 * math.Sqrt(c1p*c2p) * math.Sin(radians(deltahp/2))
	lBarp := (target.L + measured.L) / 2
	cBarp := (c1p + c2p) / 2
	var hBarp float64
	switch {
	case c1p*c2p == 0:
		hBarp = h1p + h2p
	case math.Abs(h1p-h2p) <= 180:
		hBarp = (h1p + h2p) / 2
	case h1p+h2p < 360:
		hBarp = (h1p + h2p + 360) / 2
	default:
		hBarp = (h1p + h2p - 360) / 2
	}
	t := 1 - 0.17*math.Cos(radians(hBarp-30)) + 0.24*math.Cos(radians(2*hBarp)) + 0.32*math.Cos(radians(3*hBarp+6)) - 0.20*math.Cos(radians(4*hBarp-63))
	deltaTheta := 30 * math.Exp(-square((hBarp-275)/25))
	cBarp7 := math.Pow(cBarp, 7)
	rc := 2 * math.Sqrt(cBarp7/(cBarp7+math.Pow(25, 7)))
	sl := 1 + 0.015*square(lBarp-50)/math.Sqrt(20+square(lBarp-50))
	sc := 1 + 0.045*cBarp
	sh := 1 + 0.015*cBarp*t
	rt := -math.Sin(radians(2*deltaTheta)) * rc
	lTerm := deltaLp / sl
	cTerm := deltaCp / sc
	hTerm := deltaHp / sh
	return math.Sqrt(lTerm*lTerm + cTerm*cTerm + hTerm*hTerm + rt*cTerm*hTerm)
}

func hueAngle(b, a float64) float64 {
	if a == 0 && b == 0 {
		return 0
	}
	angle := math.Atan2(b, a) * 180 / math.Pi
	if angle < 0 {
		angle += 360
	}
	return angle
}

func radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func square(value float64) float64 {
	return value * value
}

type ColorTarget struct {
	Ink       string       `json:"ink"`
	Lab       LabColor     `json:"lab"`
	Tolerance float64      `json:"tolerance"`
	Formula   ColorFormula `json:"formula"`
}

func (target ColorTarget) Validate() error {
	if target.Ink == "" {
		return errors.New("color target ink is required")
	}
	if err := target.Lab.Validate(); err != nil {
		return err
	}
	if target.Tolerance <= 0 {
		return errors.New("color target tolerance is required")
	}
	if !target.Formula.Valid() {
		return errors.New("color target formula is invalid")
	}
	return nil
}

type ColorBaseline struct {
	JobGeneration   uint64      `json:"job_generation"`
	PlateGeneration uint64      `json:"plate_generation"`
	InkBatch        string      `json:"ink_batch"`
	Target          ColorTarget `json:"target"`
	EstablishedAt   time.Time   `json:"established_at"`
}

func (baseline ColorBaseline) Validate() error {
	if baseline.JobGeneration == 0 || baseline.PlateGeneration == 0 {
		return errors.New("color baseline generation is incomplete")
	}
	if baseline.InkBatch == "" {
		return errors.New("color baseline ink batch is required")
	}
	if baseline.EstablishedAt.IsZero() {
		return errors.New("color baseline timestamp is required")
	}
	return baseline.Target.Validate()
}

func (baseline ColorBaseline) Accepts(reading ColorReading) bool {
	return reading.BelongsTo(baseline.JobGeneration, baseline.PlateGeneration, baseline.InkBatch)
}

type ColorReading struct {
	ID              string    `json:"id"`
	JobGeneration   uint64    `json:"job_generation"`
	PlateGeneration uint64    `json:"plate_generation"`
	InkBatch        string    `json:"ink_batch"`
	WebMeter        float64   `json:"web_meter"`
	Lab             LabColor  `json:"lab"`
	MeasuredAt      time.Time `json:"measured_at"`
}

func (reading ColorReading) Validate() error {
	if reading.ID == "" || reading.InkBatch == "" {
		return errors.New("color reading identity is incomplete")
	}
	if reading.JobGeneration == 0 || reading.PlateGeneration == 0 {
		return errors.New("color reading generation is incomplete")
	}
	if reading.WebMeter < 0 {
		return errors.New("color reading meter is invalid")
	}
	if err := reading.Lab.Validate(); err != nil {
		return err
	}
	if reading.MeasuredAt.IsZero() {
		return errors.New("color reading timestamp is required")
	}
	return nil
}

func (reading ColorReading) BelongsTo(jobGeneration, plateGeneration uint64, inkBatch string) bool {
	return reading.JobGeneration == jobGeneration && reading.PlateGeneration == plateGeneration && reading.InkBatch == inkBatch
}

type ColorDeviation struct {
	ReadingID string  `json:"reading_id"`
	WebMeter  float64 `json:"web_meter"`
	DeltaE    float64 `json:"delta_e"`
	Limit     float64 `json:"limit"`
	Exceeds   bool    `json:"exceeds"`
}

type ColorSegment struct {
	StartMeter float64 `json:"start_meter"`
	EndMeter   float64 `json:"end_meter"`
	Readings   int     `json:"readings"`
	MaxDeltaE  float64 `json:"max_delta_e"`
}
