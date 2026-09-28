package color

import (
	"fmt"
	"math"

	"github.com/wyw14/solo-038/internal/model"
)

const (
	FormulaCIE76     = "cie76"
	FormulaCIEDE2000 = "ciede2000"
)

func DeltaE(formula string, target, measured model.Lab) (float64, error) {
	switch formula {
	case FormulaCIE76:
		return DeltaE76(target, measured), nil
	case FormulaCIEDE2000:
		return DeltaE2000(target, measured), nil
	default:
		return 0, fmt.Errorf("unsupported delta e formula %q", formula)
	}
}

func DeltaE76(target, measured model.Lab) float64 {
	return math.Hypot(target.L-measured.L, math.Hypot(target.A-measured.A, target.B-measured.B))
}

func DeltaE2000(target, measured model.Lab) float64 {
	chroma1 := math.Hypot(target.A, target.B)
	chroma2 := math.Hypot(measured.A, measured.B)
	chromaMean := (chroma1 + chroma2) / 2
	chromaMean7 := math.Pow(chromaMean, 7)
	factor := 0.5 * (1 - math.Sqrt(chromaMean7/(chromaMean7+math.Pow(25, 7))))
	a1 := target.A * (1 + factor)
	a2 := measured.A * (1 + factor)
	c1 := math.Hypot(a1, target.B)
	c2 := math.Hypot(a2, measured.B)
	h1 := hueDegrees(a1, target.B)
	h2 := hueDegrees(a2, measured.B)
	deltaLightness := measured.L - target.L
	deltaChroma := c2 - c1
	var deltaHueAngle float64
	if c1*c2 != 0 {
		deltaHueAngle = h2 - h1
		if deltaHueAngle > 180 {
			deltaHueAngle -= 360
		} else if deltaHueAngle < -180 {
			deltaHueAngle += 360
		}
	}
	deltaHue := 2 * math.Sqrt(c1*c2) * math.Sin(radians(deltaHueAngle/2))
	lightnessMean := (target.L + measured.L) / 2
	chromaMeanPrime := (c1 + c2) / 2
	var hueMean float64
	switch {
	case c1*c2 == 0:
		hueMean = h1 + h2
	case math.Abs(h1-h2) <= 180:
		hueMean = (h1 + h2) / 2
	case h1+h2 < 360:
		hueMean = (h1 + h2 + 360) / 2
	default:
		hueMean = (h1 + h2 - 360) / 2
	}
	t := 1 - 0.17*cosDegrees(hueMean-30) + 0.24*cosDegrees(2*hueMean) + 0.32*cosDegrees(3*hueMean+6) - 0.20*cosDegrees(4*hueMean-63)
	deltaTheta := 30 * math.Exp(-math.Pow((hueMean-275)/25, 2))
	chromaMeanPrime7 := math.Pow(chromaMeanPrime, 7)
	rotation := 2 * math.Sqrt(chromaMeanPrime7/(chromaMeanPrime7+math.Pow(25, 7)))
	weightLightness := 1 + 0.015*math.Pow(lightnessMean-50, 2)/math.Sqrt(20+math.Pow(lightnessMean-50, 2))
	weightChroma := 1 + 0.045*chromaMeanPrime
	weightHue := 1 + 0.015*chromaMeanPrime*t
	rotationTerm := -sinDegrees(2*deltaTheta) * rotation
	return math.Sqrt(math.Pow(deltaLightness/weightLightness, 2) + math.Pow(deltaChroma/weightChroma, 2) + math.Pow(deltaHue/weightHue, 2) + rotationTerm*(deltaChroma/weightChroma)*(deltaHue/weightHue))
}

func hueDegrees(a, b float64) float64 {
	if a == 0 && b == 0 {
		return 0
	}
	hue := degrees(math.Atan2(b, a))
	if hue < 0 {
		hue += 360
	}
	return hue
}

func radians(value float64) float64 {
	return value * math.Pi / 180
}

func degrees(value float64) float64 {
	return value * 180 / math.Pi
}

func cosDegrees(value float64) float64 {
	return math.Cos(radians(value))
}

func sinDegrees(value float64) float64 {
	return math.Sin(radians(value))
}
