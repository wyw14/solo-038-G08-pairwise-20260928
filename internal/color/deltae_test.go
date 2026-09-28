package color

import (
	"math"
	"testing"

	"github.com/wyw14/solo-038/internal/model"
)

var sharmaPairs = []struct {
	target   model.LabColor
	measured model.LabColor
	expected float64
}{
	{model.LabColor{L: 50.0000, A: 2.6772, B: -79.7751}, model.LabColor{L: 50.0000, A: 0.0000, B: -82.7485}, 2.0425},
	{model.LabColor{L: 50.0000, A: 3.1571, B: -77.2803}, model.LabColor{L: 50.0000, A: 0.0000, B: -82.7485}, 2.8615},
	{model.LabColor{L: 50.0000, A: 2.8361, B: -74.0200}, model.LabColor{L: 50.0000, A: 0.0000, B: -82.7485}, 3.4412},
	{model.LabColor{L: 50.0000, A: -1.3802, B: -84.2814}, model.LabColor{L: 50.0000, A: 0.0000, B: -82.7485}, 1.0000},
	{model.LabColor{L: 50.0000, A: -0.9009, B: -85.5211}, model.LabColor{L: 50.0000, A: 0.0000, B: -82.7485}, 1.0000},
	{model.LabColor{L: 50.0000, A: 0.0000, B: 0.0000}, model.LabColor{L: 50.0000, A: -1.0000, B: 2.0000}, 2.3669},
	{model.LabColor{L: 50.0000, A: 2.4900, B: -0.0010}, model.LabColor{L: 50.0000, A: -2.4900, B: 0.0009}, 7.1792},
	{model.LabColor{L: 50.0000, A: 2.4900, B: -0.0010}, model.LabColor{L: 50.0000, A: -2.4900, B: 0.0011}, 7.2195},
	{model.LabColor{L: 50.0000, A: -0.0010, B: 2.4900}, model.LabColor{L: 50.0000, A: 0.0009, B: -2.4900}, 4.8045},
	{model.LabColor{L: 50.0000, A: -0.0010, B: 2.4900}, model.LabColor{L: 50.0000, A: 0.0011, B: -2.4900}, 4.7461},
	{model.LabColor{L: 50.0000, A: 2.5000, B: 0.0000}, model.LabColor{L: 50.0000, A: 0.0000, B: -2.5000}, 4.3065},
	{model.LabColor{L: 50.0000, A: 2.5000, B: 0.0000}, model.LabColor{L: 73.0000, A: 25.0000, B: -18.0000}, 27.1492},
	{model.LabColor{L: 50.0000, A: 2.5000, B: 0.0000}, model.LabColor{L: 61.0000, A: -5.0000, B: 29.0000}, 22.8977},
	{model.LabColor{L: 50.0000, A: 2.5000, B: 0.0000}, model.LabColor{L: 56.0000, A: -27.0000, B: -3.0000}, 31.9030},
	{model.LabColor{L: 50.0000, A: 2.5000, B: 0.0000}, model.LabColor{L: 58.0000, A: 24.0000, B: 15.0000}, 19.4535},
	{model.LabColor{L: 50.0000, A: 2.5000, B: 0.0000}, model.LabColor{L: 50.0000, A: 3.1736, B: 0.5854}, 1.0000},
	{model.LabColor{L: 50.0000, A: 2.5000, B: 0.0000}, model.LabColor{L: 50.0000, A: 1.8634, B: 0.5757}, 1.0000},
	{model.LabColor{L: 60.2574, A: -34.0099, B: 36.2677}, model.LabColor{L: 60.4626, A: -34.1751, B: 39.4387}, 1.2644},
	{model.LabColor{L: 22.7233, A: 20.0904, B: -46.6940}, model.LabColor{L: 23.0331, A: 14.9730, B: -42.5619}, 2.0373},
	{model.LabColor{L: 2.0776, A: 0.0795, B: -1.1350}, model.LabColor{L: 0.9033, A: -0.0636, B: -0.5514}, 0.9082},
}

func TestDeltaE2000SharmaPairs(t *testing.T) {
	for index, pair := range sharmaPairs {
		actual := model.DeltaE2000(pair.target, pair.measured)
		if math.Abs(actual-pair.expected) > 0.0001 {
			t.Errorf("pair %d: expected %.4f, got %.4f", index, pair.expected, actual)
		}
	}
}

func TestDeltaE76(t *testing.T) {
	target := model.LabColor{L: 50, A: 2.5, B: 0}
	measured := model.LabColor{L: 50, A: 2.5, B: 0}
	if delta := model.DeltaE76(target, measured); delta != 0 {
		t.Errorf("identical colors: expected 0, got %.4f", delta)
	}
	measured = model.LabColor{L: 53, A: 6.5, B: 0}
	if delta := model.DeltaE76(target, measured); delta != 5 {
		t.Errorf("expected 5, got %.4f", delta)
	}
}

func TestColorFormulaDispatch(t *testing.T) {
	target := model.LabColor{L: 50, A: 2.5, B: 0}
	measured := model.LabColor{L: 50, A: 3.1736, B: 0.5854}
	if delta := model.ColorFormulaCIE2000.DeltaE(target, measured); math.Abs(delta-1.0) > 0.0001 {
		t.Errorf("cie2000 dispatch: expected 1.0, got %.4f", delta)
	}
	if delta := model.ColorFormulaCIE76.DeltaE(target, measured); delta != model.DeltaE76(target, measured) {
		t.Errorf("cie76 dispatch mismatch: %.6f vs %.6f", delta, model.DeltaE76(target, measured))
	}
}

func TestMergeSegments(t *testing.T) {
	deviations := []model.ColorDeviation{
		{ReadingID: "a", WebMeter: 10, DeltaE: 1.0, Limit: 2, Exceeds: false},
		{ReadingID: "b", WebMeter: 20, DeltaE: 2.5, Limit: 2, Exceeds: true},
		{ReadingID: "c", WebMeter: 30, DeltaE: 3.1, Limit: 2, Exceeds: true},
		{ReadingID: "d", WebMeter: 40, DeltaE: 0.4, Limit: 2, Exceeds: false},
		{ReadingID: "e", WebMeter: 50, DeltaE: 2.2, Limit: 2, Exceeds: true},
	}
	segments := MergeSegments(deviations)
	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}
	if segments[0].StartMeter != 20 || segments[0].EndMeter != 30 || segments[0].Readings != 2 || segments[0].MaxDeltaE != 3.1 {
		t.Errorf("first segment is wrong: %+v", segments[0])
	}
	if segments[1].StartMeter != 50 || segments[1].EndMeter != 50 || segments[1].Readings != 1 {
		t.Errorf("second segment is wrong: %+v", segments[1])
	}
	if segments := MergeSegments(deviations[:1]); len(segments) != 0 {
		t.Errorf("expected no segments, got %d", len(segments))
	}
}
