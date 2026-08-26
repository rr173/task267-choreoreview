package beat

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestBeatAlignment(t *testing.T) {
	anchors := []*model.BeatAnchor{
		{BeatNo: 1, ImageTime: 0.0, Confidence: 0.9},
		{BeatNo: 2, ImageTime: 0.5, Confidence: 0.9},
		{BeatNo: 3, ImageTime: 1.0, Confidence: 0.9},
		{BeatNo: 4, ImageTime: 1.5, Confidence: 0.9},
	}
	res := New().Align(anchors)
	if res.MatchedCount != 4 || res.BeatCount != 4 {
		t.Fatalf("matched=%d beat_count=%d", res.MatchedCount, res.BeatCount)
	}
	if res.MaxOffsetMs > 1.0 {
		t.Fatalf("expected near-zero offsets, got %v", res.MaxOffsetMs)
	}
}

func TestBeatAlignmentMedianTempo(t *testing.T) {
	anchors := []*model.BeatAnchor{
		{BeatNo: 1, ImageTime: 0.0, Confidence: 0.9},
		{BeatNo: 2, ImageTime: 0.5, Confidence: 0.9},
		{BeatNo: 4, ImageTime: 2.5, Confidence: 0.9}, // outlier gap should not dominate
		{BeatNo: 5, ImageTime: 3.0, Confidence: 0.9},
	}
	res := New().Align(anchors)
	if res.BeatCount != 5 {
		t.Fatalf("beat_count=%d", res.BeatCount)
	}
	if res.Offsets[2].Anchored {
		t.Fatal("beat 3 should be unanchored")
	}
}

func TestBeatAlignmentSparseAnchors(t *testing.T) {
	anchors := []*model.BeatAnchor{
		{BeatNo: 1, ImageTime: 0.0, Confidence: 0.9},
		{BeatNo: 2, ImageTime: 0.5, Confidence: 0.9},
		{BeatNo: 4, ImageTime: 1.5, Confidence: 0.9},
		{BeatNo: 6, ImageTime: 2.5, Confidence: 0.9},
	}
	res := New().Align(anchors)
	if res.BeatCount != 6 || res.MatchedCount != 4 {
		t.Fatalf("beat_count=%d matched=%d", res.BeatCount, res.MatchedCount)
	}
	if res.Offsets[2].Anchored {
		t.Fatal("beat 3 should be unanchored in sparse layout")
	}
}
