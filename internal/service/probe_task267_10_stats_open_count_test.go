package service

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestStatsOpenVariantCountAfterAdjudication(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("stats", "Yunnan", 2, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Movement().AddMovement(d.ID, 1, 1, 2, "a", model.ConnectionContinuous); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Movement().AddMovement(d.ID, 1, 5, 6, "b", model.ConnectionContinuous); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Formation().AddFormation(d.ID, 1, 2, 1, 4, "line"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Formation().AddFormation(d.ID, 2, 1, 1, 4, "diagonal"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Analyze().RunFullAnalysis(d.ID); err != nil {
		t.Fatal(err)
	}
	open, err := svc.Variant().OpenVariants(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range open {
		if _, err := svc.Variant().Adjudicate(v.ID, model.VariantStatusRejected, "test"); err != nil {
			t.Fatal(err)
		}
	}
	stats, err := svc.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.OpenVariantCount != 0 {
		t.Fatalf("open_variant_count=%d want 0", stats.OpenVariantCount)
	}
}
