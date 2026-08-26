package service

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestAdjudicateFormationUpdatesEdgeStatus(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("form-sync", "Yunnan", 2, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Formation().AddFormation(d.ID, 1, 2, 1, 4, "line"); err != nil {
		t.Fatal(err)
	}
	f2, err := svc.Formation().AddFormation(d.ID, 2, 1, 1, 4, "diagonal")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Analyze().RunFullAnalysis(d.ID); err != nil {
		t.Fatal(err)
	}
	open, err := svc.Variant().OpenVariants(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	var formVar *model.VariantCandidate
	for _, v := range open {
		if v.Type == model.VariantTypeFormation {
			formVar = v
			break
		}
	}
	if formVar == nil {
		t.Fatal("missing formation variant")
	}
	if _, err := svc.Variant().Adjudicate(formVar.ID, model.VariantStatusConfirmed, "local"); err != nil {
		t.Fatal(err)
	}
	edges, err := svc.Formation().ListFormations(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range edges {
		if e.ID == f2.ID && e.Status != model.FormationStatusConfirmed {
			t.Fatalf("adjudicated formation edge status=%s", e.Status)
		}
	}
}
