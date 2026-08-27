package service

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestDraftVersionCannotFreezeDirectly(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("ver-guard", "Yunnan", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, to := range []string{model.DanceStatusAligning, model.DanceStatusReviewing, model.DanceStatusPublished} {
		if _, err := svc.Dance().TransitDance(d.ID, to); err != nil {
			t.Fatal(err)
		}
	}
	ver, err := svc.Version().CreateVersion(d.ID, "v1", "evidence")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Version().FreezeVersion(ver.ID); err == nil {
		t.Fatal("draft version must not freeze directly")
	}
}
