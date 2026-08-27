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

// TestBeatAlignmentOutlierGapRobust 校验节奏估算对离群间隔的鲁棒性：
// 以 0.5 秒/拍为真实节奏的序列里混入一段被拉长到 5 秒/拍的异常间隔（beat 5→6）。
// 中位数估算法应仍收敛到 0.5，离群点之前的稳定节拍偏移保持毫秒量级；
// 而均值估算法会把节奏拉偏到 ~1.25 秒/拍，使 beat 2..5 的偏移系统性漂移到数百毫秒。
func TestBeatAlignmentOutlierGapRobust(t *testing.T) {
	anchors := []*model.BeatAnchor{
		{BeatNo: 1, ImageTime: 0.0, Confidence: 0.9},
		{BeatNo: 2, ImageTime: 0.5, Confidence: 0.9},
		{BeatNo: 3, ImageTime: 1.0, Confidence: 0.9},
		{BeatNo: 4, ImageTime: 1.5, Confidence: 0.9},
		{BeatNo: 5, ImageTime: 2.0, Confidence: 0.9},
		// 离群间隔：到第 6 拍本应 ~2.5s，却被拉到 7.5s（5 秒/拍）。
		{BeatNo: 6, ImageTime: 7.5, Confidence: 0.9},
		{BeatNo: 7, ImageTime: 8.0, Confidence: 0.9},
	}
	res := New().Align(anchors)
	if res.MatchedCount != len(anchors) {
		t.Fatalf("matched=%d want=%d", res.MatchedCount, len(anchors))
	}
	// beat 1..5 落在稳定节奏栅格上，偏移应保持在毫秒量级，不被离群间隔带偏。
	for b := 1; b <= 5; b++ {
		off := res.Offsets[b-1]
		if !off.Anchored {
			t.Fatalf("beat %d should be anchored", b)
		}
		abs := off.OffsetMs
		if abs < 0 {
			abs = -abs
		}
		if abs > 1.0 {
			t.Errorf("beat %d offset_ms=%.3f drifts beyond 1ms; tempo not robust to outlier gap", b, off.OffsetMs)
		}
	}
}
