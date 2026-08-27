package service

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestDanceCannotSkipLifecycleStates(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("life", "Yunnan", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Dance().TransitDance(d.ID, model.DanceStatusPublished); err == nil {
		t.Fatal("organizing must not jump to published")
	}
}
