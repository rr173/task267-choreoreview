package service

import (
	"testing"

	"task267-choreoreview/internal/variant"
)

func TestSparseAnchorMiddleGapMaterializesBeatCandidate(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("sparse-beat", "Yunnan", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []struct{ beat int; t float64 }{
		{1, 0.0}, {2, 0.5}, {4, 1.5}, {6, 2.5},
	} {
		if _, err := svc.Beat().AddBeatAnchor(d.ID, in.beat, in.t, 0.9, "cam"); err != nil {
			t.Fatal(err)
		}
	}
	align, err := svc.Analyze().AlignBeats(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	cands := variant.New().BuildBeatCandidates(d.ID, align, 250.0)
	var beat3 bool
	for _, c := range cands {
		if c.Type == "beat" && c.RefID == 3 {
			beat3 = true
		}
	}
	if !beat3 {
		t.Fatalf("expected beat-3 candidate for sparse middle gap, got %+v", cands)
	}
}
