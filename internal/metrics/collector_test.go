package metrics

import (
	"testing"
	"time"
)

func TestAppendPointCapsSize(t *testing.T) {
	points := []Point{}
	for i := range 5 {
		points = appendPoint(points, Point{TS: time.Now(), Val: float64(i)}, 3)
	}
	if len(points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(points))
	}
	if points[0].Val != 2 || points[2].Val != 4 {
		t.Fatalf("unexpected points after cap: %#v", points)
	}
}
