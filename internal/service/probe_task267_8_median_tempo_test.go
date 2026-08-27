package service

import (
	"math"
	"testing"
)

func TestOutlierGapDoesNotSkewBeatOffsets(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("tempo", "Yunnan", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []struct{ beat int; t float64 }{
		{1, 0.0}, {2, 0.5}, {3, 1.0}, {4, 1.5}, {6, 3.0},
	} {
		if _, err := svc.Beat().AddBeatAnchor(d.ID, in.beat, in.t, 0.9, "cam"); err != nil {
			t.Fatal(err)
		}
	}
	align, err := svc.Analyze().AlignBeats(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(align.Offsets) < 4 {
		t.Fatal("expected offsets")
	}
	off := math.Abs(align.Offsets[3].OffsetMs)
	if off > 100 {
		t.Fatalf("beat 4 offset too large with median tempo: %v", align.Offsets[3].OffsetMs)
	}
}
