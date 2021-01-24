package main

import (
	"math"
	"testing"
)

const floatAlmostEqualDiff = 1e-9

func compareFloats(a, b float64) bool {
	return math.Abs(a-b) < floatAlmostEqualDiff
}

func TestPressureFromVolumes(t *testing.T) {
	cylinderVolume := CylinderVolume(12)
	gasVolume := GasVolume(12)
	pressure := PressureFromVolumes(gasVolume, cylinderVolume)
	if pressure != 1.0 {
		t.Errorf("Invalid pressure, expected 1.0, got %f", pressure)
	}
	gasVolume = GasVolume(2000)
	pressure = PressureFromVolumes(gasVolume, cylinderVolume)
	expectedValue := 166.0 + 2.0/3.0
	if !compareFloats(float64(pressure), expectedValue) {
		t.Errorf("Invalid pressure, expected %f, got %f", expectedValue, pressure)
	}
}

func TestPartialPressure(t *testing.T) {
	pressure := PressureBar(166.0 + 2.0/3.0)
	partialPressure := pressure.PartialPressure(0.21)
	if !compareFloats(float64(partialPressure), 35.0) {
		t.Errorf("Invalid pressure, expected 35, got %f", partialPressure)

	}
}

func TestGasWeightFromMole(t *testing.T) {
	atomicWeight := AtomicWeightLookup[Argon]
	moleCount := MoleCount(32.5)
	weight := GasWeightFromMole(moleCount, atomicWeight)
	expectedWeight := 455.217750
	if !compareFloats(float64(weight), expectedWeight) {
		t.Errorf("Invalid gas weight %f, expected %f", weight, expectedWeight)
	}
}
