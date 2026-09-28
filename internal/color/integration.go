package color

import (
	"sort"

	"github.com/wyw14/solo-038/internal/model"
)

func BuildSegments(measurements []model.Measurement) []model.Segment {
	ordered := append([]model.Measurement(nil), measurements...)
	sort.Slice(ordered, func(left, right int) bool { return ordered[left].WebMeter < ordered[right].WebMeter })
	segments := []model.Segment{}
	open := -1
	for _, measurement := range ordered {
		if measurement.Within {
			open = -1
			continue
		}
		if open < 0 {
			segments = append(segments, model.Segment{StartMeter: measurement.WebMeter, MeasurementIDs: []string{}})
			open = len(segments) - 1
		}
		segment := &segments[open]
		segment.EndMeter = measurement.WebMeter
		segment.Samples++
		if measurement.DeltaE > segment.MaxDeltaE {
			segment.MaxDeltaE = measurement.DeltaE
		}
		segment.MeasurementIDs = append(segment.MeasurementIDs, measurement.ID)
	}
	return segments
}

func meterSpan(measurements []model.Measurement) (float64, float64) {
	start := measurements[0].WebMeter
	end := measurements[0].WebMeter
	for _, measurement := range measurements[1:] {
		if measurement.WebMeter < start {
			start = measurement.WebMeter
		}
		if measurement.WebMeter > end {
			end = measurement.WebMeter
		}
	}
	return start, end
}

func measurementIDs(measurements []model.Measurement) []string {
	ids := make([]string, 0, len(measurements))
	for _, measurement := range measurements {
		ids = append(ids, measurement.ID)
	}
	sort.Strings(ids)
	return ids
}
