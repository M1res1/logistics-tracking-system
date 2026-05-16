package service

import (
	"math"
	"testing"
)

func TestHaversineDistance(t *testing.T) {
	dist := HaversineDistance(40.7128, -74.0060, 34.0522, -118.2437)
	if math.Abs(dist-3940) > 50 {
		t.Errorf("NY to LA: expected ~3940 km, got %.2f", dist)
	}

	dist = HaversineDistance(51.5, -0.1, 51.5, -0.1)
	if dist > 0.001 {
		t.Errorf("same point: expected 0, got %.6f", dist)
	}

	dist = HaversineDistance(51.5074, -0.1278, 48.8566, 2.3522)
	if math.Abs(dist-341) > 5 {
		t.Errorf("London to Paris: expected ~341 km, got %.2f", dist)
	}
}

func TestEstimateETA(t *testing.T) {
	if eta := EstimateETA(30); eta != 60 {
		t.Errorf("30km: expected 60 min, got %d", eta)
	}
	if eta := EstimateETA(15); eta != 30 {
		t.Errorf("15km: expected 30 min, got %d", eta)
	}
	if eta := EstimateETA(1); eta != 2 {
		t.Errorf("1km: expected 2 min, got %d", eta)
	}
}
