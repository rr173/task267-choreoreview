package service

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestSuccessorJumpNotFlaggedAsBreak(t *testing.T) {
	svc := newTestService(t)
	d, err := svc.Dance().CreateDance("jump", "Yunnan", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Movement().AddMovement(d.ID, 1, 1, 2, "step", model.ConnectionContinuous); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Movement().AddMovement(d.ID, 1, 7, 8, "leap", model.ConnectionJump); err != nil {
		t.Fatal(err)
	}
	check, err := svc.Analyze().CheckContinuity(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if check.BrokenCount != 0 {
		t.Fatalf("successor jump must not create break: %+v", check.Broken)
	}
}
