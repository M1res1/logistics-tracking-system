package repository

import (
	"math"
	"testing"
)

// TestHaversineKm validates the haversine distance formula with known coordinate pairs.
func TestHaversineKm(t *testing.T) {
	cases := []struct {
		name        string
		lat1, lng1  float64
		lat2, lng2  float64
		expectedKm  float64
		toleranceKm float64
	}{
		{
			name:        "same point",
			lat1: 40.7128, lng1: -74.0060,
			lat2: 40.7128, lng2: -74.0060,
			expectedKm: 0, toleranceKm: 0.001,
		},
		{
			name:        "New York to Los Angeles",
			lat1: 40.7128, lng1: -74.0060,
			lat2: 34.0522, lng2: -118.2437,
			expectedKm: 3940, toleranceKm: 50,
		},
		{
			name:        "London to Paris",
			lat1: 51.5074, lng1: -0.1278,
			lat2: 48.8566, lng2: 2.3522,
			expectedKm: 341, toleranceKm: 5,
		},
		{
			name:        "nearby points within 1 km",
			lat1: 41.0, lng1: 29.0,
			lat2: 41.005, lng2: 29.005,
			expectedKm: 0.65, toleranceKm: 0.1,
		},
		{
			name:        "Istanbul to Ankara",
			lat1: 41.0082, lng1: 28.9784,
			lat2: 39.9208, lng2: 32.8541,
			expectedKm: 352, toleranceKm: 10,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := haversineKm(tc.lat1, tc.lng1, tc.lat2, tc.lng2)
			if math.Abs(got-tc.expectedKm) > tc.toleranceKm {
				t.Errorf("haversineKm(%v,%v → %v,%v) = %.4f km; want %.4f ±%.4f",
					tc.lat1, tc.lng1, tc.lat2, tc.lng2, got, tc.expectedKm, tc.toleranceKm)
			}
		})
	}
}

// TestHaversineKm_Symmetry verifies distance(A,B) == distance(B,A).
func TestHaversineKm_Symmetry(t *testing.T) {
	d1 := haversineKm(48.8566, 2.3522, 51.5074, -0.1278)
	d2 := haversineKm(51.5074, -0.1278, 48.8566, 2.3522)
	if math.Abs(d1-d2) > 0.001 {
		t.Errorf("haversine should be symmetric: got %.6f and %.6f", d1, d2)
	}
}
