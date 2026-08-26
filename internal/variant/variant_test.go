package variant

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestBuildMovementCandidates(t *testing.T) {
	cont := &model.ContinuityCheck{
		BrokenCount: 1,
		Broken: []model.BrokenSegment{{
			MovementID: 10, DancerNo: 2, ActionName: "大崴", EndBeat: 2, NextAction: "颠步", NextBeat: 5, GapBeats: 2,
		}},
	}
	cands := New().BuildMovementCandidates(7, cont)
	if len(cands) != 1 || cands[0].Type != model.VariantTypeMovement {
		t.Fatalf("candidates=%+v", cands)
	}
}

func TestBuildBeatCandidatesMissingAnchors(t *testing.T) {
	align := &model.AlignmentResult{
		BeatCount:    6,
		MatchedCount: 4,
		Offsets: []model.BeatOffset{
			{BeatNo: 1, Anchored: true},
			{BeatNo: 2, Anchored: true},
			{BeatNo: 3, Anchored: true},
			{BeatNo: 4, Anchored: true},
			{BeatNo: 5, Anchored: false},
			{BeatNo: 6, Anchored: false},
		},
	}
	cands := New().BuildBeatCandidates(1, align, 250.0)
	if len(cands) != 2 {
		t.Fatalf("candidates=%d want 2", len(cands))
	}
}

func TestBuildBeatCandidatesSparseMiddleGap(t *testing.T) {
	align := &model.AlignmentResult{
		BeatCount:    6,
		MatchedCount: 4,
		Offsets: []model.BeatOffset{
			{BeatNo: 1, Anchored: true},
			{BeatNo: 2, Anchored: true},
			{BeatNo: 3, Anchored: false},
			{BeatNo: 4, Anchored: true},
			{BeatNo: 5, Anchored: false},
			{BeatNo: 6, Anchored: true},
		},
	}
	cands := New().BuildBeatCandidates(1, align, 250.0)
	if len(cands) != 2 {
		t.Fatalf("expected beat 3 and 5 candidates, got %d", len(cands))
	}
}
