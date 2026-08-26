// 节拍模块：谱记节拍与影像锚点的节奏对齐。
package beat

import (
	"sort"
	"time"

	"task267-choreoreview/internal/model"
)

// Tempo 节拍对齐的节奏参数。
type Tempo struct {
	// BeatsPerMinute 依据锚点对推断的节奏。
	BeatsPerMinute float64
	// FirstBeatOffsetMs 首个锚点相对谱记节拍 1 的影像偏移（毫秒）。
	FirstBeatOffsetMs float64
}

// Service 节拍对齐服务。
type Service struct{}

// New 构造节拍服务。
func New() *Service {
	return &Service{}
}

// Align 将谱记节拍与影像锚点对齐：
//   - 对每个锚点，谱记节拍号与影像时间构成 (beatNo, imageTime) 点对；
//   - 用相邻锚点对估算节奏（秒/拍）；
//   - 对每个谱记节拍计算预期影像时间，与锚点比对得到偏移（毫秒）。
func (s *Service) Align(anchors []*model.BeatAnchor) *model.AlignmentResult {
	res := &model.AlignmentResult{
		AnchorCount: len(anchors),
		Offsets:     []model.BeatOffset{},
		GeneratedAt: time.Now().UTC(),
	}
	if len(anchors) == 0 {
		return res
	}

	// 按节拍号排序。
	sorted := append([]*model.BeatAnchor(nil), anchors...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].BeatNo < sorted[j].BeatNo })

	// 最大谱记节拍号。
	maxBeat := 0
	for _, a := range sorted {
		if a.BeatNo > maxBeat {
			maxBeat = a.BeatNo
		}
	}
	res.BeatCount = maxBeat

	// 估算节奏：优先用前两个相邻锚点对的中位数秒差。
	secPerBeat := estimateTempo(sorted)

	// 计算参考零拍（第一个锚点）的影像时间。
	first := sorted[0]
	refTime := first.ImageTime - float64(first.BeatNo-1)*secPerBeat

	// 对每个谱记节拍，计算预期时间并匹配锚点。
	anchorByBeat := map[int]*model.BeatAnchor{}
	for _, a := range sorted {
		anchorByBeat[a.BeatNo] = a
	}

	totalOffset := 0.0
	matched := 0
	for beat := 1; beat <= maxBeat; beat++ {
		expected := refTime + float64(beat-1)*secPerBeat
		off := model.BeatOffset{BeatNo: beat, ImageTime: expected}
		if a, ok := anchorByBeat[beat]; ok {
			off.ImageTime = a.ImageTime
			off.OffsetMs = (a.ImageTime - expected) * 1000
			off.Anchored = true
			totalOffset += off.OffsetMs
			matched++
		}
		res.Offsets = append(res.Offsets, off)
	}
	res.MatchedCount = matched
	if matched > 0 {
		res.AvgOffsetMs = totalOffset / float64(matched)
	}
	for _, off := range res.Offsets {
		if off.Anchored {
			a := off.OffsetMs
			if a < 0 {
				a = -a
			}
			if a > res.MaxOffsetMs {
				res.MaxOffsetMs = a
			}
		}
	}
	return res
}

// estimateTempo 用相邻锚点间隔的中位秒差估算每拍秒数。
func estimateTempo(sorted []*model.BeatAnchor) float64 {
	if len(sorted) < 2 {
		return 0.5 // 默认 120 BPM
	}
	var gaps []float64
	for i := 1; i < len(sorted); i++ {
		beatGap := sorted[i].BeatNo - sorted[i-1].BeatNo
		if beatGap > 0 {
			gaps = append(gaps, (sorted[i].ImageTime-sorted[i-1].ImageTime)/float64(beatGap))
		}
	}
	if len(gaps) == 0 {
		return 0.5
	}
	var sum float64
	for _, g := range gaps {
		sum += g
	}
	return sum / float64(len(gaps))
}
