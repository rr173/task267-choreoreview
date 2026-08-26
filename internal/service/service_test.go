package service

import (
	"path/filepath"
	"testing"

	"task267-choreoreview/internal/model"
	"task267-choreoreview/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	path := filepath.Join(t.TempDir(), "svc.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return New(st)
}

func TestRunFullAnalysisMaterializesCandidates(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("test-dance", "Yunnan", 2, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Movement().AddMovement(d.ID, 1, 1, 2, "step-a", model.ConnectionContinuous); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Movement().AddMovement(d.ID, 1, 5, 6, "step-b", model.ConnectionContinuous); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Formation().AddFormation(d.ID, 1, 2, 1, 4, "line"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Formation().AddFormation(d.ID, 2, 1, 1, 4, "diagonal"); err != nil {
		t.Fatal(err)
	}
	res, err := svc.Analyze().RunFullAnalysis(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.Continuity.BrokenCount == 0 || res.Formation.ConflictCount == 0 {
		t.Fatalf("expected findings, continuity=%d formation=%d", res.Continuity.BrokenCount, res.Formation.ConflictCount)
	}
	if len(res.OpenCandidates) == 0 {
		t.Fatal("expected open candidates")
	}
}

func TestRunFullAnalysisIdempotent(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("idem-dance", "Yunnan", 1, "")
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
		t.Fatalf("second analyze should be idempotent: %v", err)
	}
}

func TestAdjudicateFormationSyncsEdge(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("form-dance", "Yunnan", 2, "")
	if err != nil {
		t.Fatal(err)
	}
	f1, err := svc.Formation().AddFormation(d.ID, 1, 2, 1, 4, "line")
	if err != nil {
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
		t.Fatal("expected formation variant")
	}
	if _, err := svc.Variant().Adjudicate(formVar.ID, model.VariantStatusConfirmed, "local variant"); err != nil {
		t.Fatal(err)
	}
	edges, err := svc.Formation().ListFormations(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range edges {
		if e.ID == f2.ID && e.Status != model.FormationStatusConfirmed {
			t.Fatalf("conflict formation edge should sync to confirmed, got %s", e.Status)
		}
		if e.ID == f1.ID && e.Status != model.FormationStatusRaw {
			t.Fatalf("line edge should stay raw, got %s", e.Status)
		}
	}
}
