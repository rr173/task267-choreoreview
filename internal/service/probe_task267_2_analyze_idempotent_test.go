package service

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestRunFullAnalysisSecondPassIdempotent(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("idem", "Yunnan", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Movement().AddMovement(d.ID, 1, 1, 2, "a", model.ConnectionContinuous); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Movement().AddMovement(d.ID, 1, 5, 6, "b", model.ConnectionContinuous); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Analyze().RunFullAnalysis(d.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Analyze().RunFullAnalysis(d.ID); err != nil {
		t.Fatalf("second analyze must stay idempotent: %v", err)
	}
}
