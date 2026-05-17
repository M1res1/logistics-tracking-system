package service

import "testing"

// TestEstimateETAZero verifies that zero distance yields zero ETA.
func TestEstimateETAZero(t *testing.T) {
	if eta := EstimateETA(0); eta != 0 {
		t.Errorf("0km: expected 0 min, got %d", eta)
	}
}

// TestHaversineSymmetry verifies that distance(A→B) == distance(B→A).
func TestHaversineSymmetry(t *testing.T) {
	tests := []struct {
		name string
		lat1, lng1 float64
		lat2, lng2 float64
	}{
		{
			name: "Tashkent to Moscow",
			lat1: 41.2995, lng1: 69.2401,
			lat2: 55.7558, lng2: 37.6173,
		},
		{
			name: "New York to London",
			lat1: 40.7128, lng1: -74.0060,
			lat2: 51.5074, lng2: -0.1278,
		},
		{
			name: "same point",
			lat1: 48.8566, lng1: 2.3522,
			lat2: 48.8566, lng2: 2.3522,
		},
		{
			name: "equator crossing",
			lat1: 1.0, lng1: 0.0,
			lat2: -1.0, lng2: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d1 := HaversineDistance(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			d2 := HaversineDistance(tt.lat2, tt.lng2, tt.lat1, tt.lng1)
			if d1 != d2 {
				t.Errorf("haversine not symmetric for %s: %.4f vs %.4f", tt.name, d1, d2)
			}
		})
	}
}
