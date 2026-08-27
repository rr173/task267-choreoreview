package service

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestAnalyzeMarksBrokenMovements(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("broken", "Yunnan", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	m1, err := svc.Movement().AddMovement(d.ID, 1, 1, 2, "a", model.ConnectionContinuous)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Movement().AddMovement(d.ID, 1, 5, 6, "b", model.ConnectionContinuous); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Analyze().RunFullAnalysis(d.ID); err != nil {
		t.Fatal(err)
	}
	mv, err := svc.Movement().ListMovements(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range mv {
		if u.ID == m1.ID && u.Status != model.MovementStatusBroken {
			t.Fatalf("broken movement should be marked broken, got %s", u.Status)
		}
	}
}
