// 裁决模块：异读候选生成与裁决语义。
package variant

import (
	"fmt"
	"strconv"
	"time"

	"task267-choreoreview/internal/model"
)

// Service 异读裁决服务：将校验结论物化为候选，并支持地方变体保留。
type Service struct{}

// New 构造裁决服务。
func New() *Service {
	return &Service{}
}

// BuildMovementCandidates 将动作断裂段物化为异读候选。
//
// 每个断裂段生成一个 movement 类候选，ref_id 指向断裂动作单元，
// detail 记录断裂位置与缺口节拍数。
func (s *Service) BuildMovementCandidates(danceID int64, check *model.ContinuityCheck) []*model.VariantCandidate {
	out := make([]*model.VariantCandidate, 0, len(check.Broken))
	for _, b := range check.Broken {
		out = append(out, &model.VariantCandidate{
			DanceID: danceID,
			Type:    model.VariantTypeMovement,
			RefID:   b.MovementID,
			Detail:  fmt.Sprintf("movement break: dancer %d %s ends at beat %d gap %d beat(s)",
				b.DancerNo, b.ActionName, b.EndBeat, b.GapBeats),
			Status: model.VariantStatusCandidate,
		})
	}
	return out
}

// BuildFormationCandidates 将队形冲突物化为异读候选。
func (s *Service) BuildFormationCandidates(danceID int64, check *model.FormationCheck) []*model.VariantCandidate {
	out := make([]*model.VariantCandidate, 0, len(check.Conflicts))
	for _, c := range check.Conflicts {
		out = append(out, &model.VariantCandidate{
			DanceID: danceID,
			Type:    model.VariantTypeFormation,
			RefID:   c.RelationID,
			Detail:  fmt.Sprintf("formation conflict: dancers %d/%d relation %s beats [%d,%d]",
				c.FromDancer, c.ToDancer, c.Relation, c.BeatStart, c.BeatEnd),
			Status: model.VariantStatusCandidate,
		})
	}
	return out
}

// BuildBeatCandidates 将节拍对齐异常（谱记节拍无影像锚点）物化为候选。
//
// 凡 Anchored=false 的谱记节拍均缺失影像锚点，无论其位于序列首段、
// 中段稀疏缺口还是末段尾部，都应物化为一个 beat 类候选。
// 注意：不能以 BeatNo 与 MatchedCount 的大小关系为门槛——MatchedCount
// 只是已锚定节拍的数量，锚点稀疏时中段未锚拍会被错误丢弃。
func (s *Service) BuildBeatCandidates(danceID int64, align *model.AlignmentResult, thresholdMs float64) []*model.VariantCandidate {
	var out []*model.VariantCandidate
	for _, off := range align.Offsets {
		if off.Anchored {
			continue
		}
		out = append(out, &model.VariantCandidate{
			DanceID: danceID,
			Type:    model.VariantTypeBeat,
			RefID:   int64(off.BeatNo),
			Detail: fmt.Sprintf("beat %d missing image anchor (threshold %s ms)",
				off.BeatNo, strconv.FormatFloat(thresholdMs, 'f', 1, 64)),
			Status: model.VariantStatusCandidate,
		})
	}
	return out
}

// LocalVariantVerdict 地方变体裁决结果（确认并保留为地方变体）。
func (s *Service) LocalVariantVerdict() string {
	return model.VariantStatusConfirmed
}

// Now 返回当前 UTC 时间（供裁决时间戳使用）。
func (s *Service) Now() time.Time {
	return time.Now().UTC()
}
