package service

import (
	"testing"
)

func TestVersionCreateRequiresPublishedDance(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("early-ver", "Yunnan", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Version().CreateVersion(d.ID, "too early", "evidence"); err == nil {
		t.Fatal("organizing dance must not create version")
	}
}
