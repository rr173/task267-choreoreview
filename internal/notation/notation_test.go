package notation

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestContinuityDetectsBreak(t *testing.T) {
	units := []*model.MovementUnit{
		{ID: 1, DancerNo: 1, ActionName: "崴步", StartBeat: 1, EndBeat: 2, Connection: model.ConnectionContinuous},
		{ID: 2, DancerNo: 1, ActionName: "颠步", StartBeat: 5, EndBeat: 6, Connection: model.ConnectionContinuous},
	}
	check := New().CheckContinuity(units)
	if !HasBreak(check) {
		t.Fatal("expected break detected")
	}
	if check.BrokenCount != 1 || check.Broken[0].GapBeats != 2 {
		t.Fatalf("broken=%+v", check.Broken)
	}
}

func TestContinuityExplicitJumpOK(t *testing.T) {
	units := []*model.MovementUnit{
		{ID: 1, DancerNo: 1, ActionName: "大跳", StartBeat: 1, EndBeat: 2, Connection: model.ConnectionJump},
		{ID: 2, DancerNo: 1, ActionName: "落步", StartBeat: 7, EndBeat: 8, Connection: model.ConnectionContinuous},
	}
	check := New().CheckContinuity(units)
	if HasBreak(check) {
		t.Fatalf("jump should not be a break: %+v", check.Broken)
	}
}

func TestContinuitySuccessorJumpNoBreak(t *testing.T) {
	units := []*model.MovementUnit{
		{ID: 1, DancerNo: 1, ActionName: "崴步", StartBeat: 1, EndBeat: 2, Connection: model.ConnectionContinuous},
		{ID: 2, DancerNo: 1, ActionName: "大跳", StartBeat: 7, EndBeat: 8, Connection: model.ConnectionJump},
	}
	check := New().CheckContinuity(units)
	if HasBreak(check) {
		t.Fatalf("successor jump should not be flagged: %+v", check.Broken)
	}
}
