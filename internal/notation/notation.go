// 动作谱记模块：动作单元导入与动作连续性校验。
package notation

import (
	"sort"
	"time"

	"task267-choreoreview/internal/model"
)

// Service 谱记模块：对动作单元序列做连续性分析。
type Service struct{}

// New 构造谱记服务。
func New() *Service {
	return &Service{}
}

// CheckContinuity 校验每个舞者的动作序列：前一动作末拍与后一动作首拍是否衔接。
//
// 规则：
//   - connection=continuous 的动作，其后一动作首拍必须等于前一动作末拍+1；
//   - 否则记为断裂（broken），给出缺口节拍数；
//   - 同一舞者的动作按 start_beat 排序后逐对检查。
func (s *Service) CheckContinuity(units []*model.MovementUnit) *model.ContinuityCheck {
	// 按舞者分组。
	byDancer := map[int][]*model.MovementUnit{}
	for _, u := range units {
		byDancer[u.DancerNo] = append(byDancer[u.DancerNo], u)
	}

	out := &model.ContinuityCheck{
		TotalUnits:  len(units),
		CheckedAt:   time.Now().UTC(),
	}

	// 每个舞者内部按 start_beat 排序。
	for _, seq := range byDancer {
		sort.Slice(seq, func(i, j int) bool {
			if seq[i].StartBeat == seq[j].StartBeat {
				return seq[i].ID < seq[j].ID
			}
			return seq[i].StartBeat < seq[j].StartBeat
		})
		for i := 0; i < len(seq); i++ {
			cur := seq[i]
			if cur.Connection == model.ConnectionJump {
				continue // 显式跳跃，不视为断裂
			}
			if i+1 < len(seq) {
				next := seq[i+1]
				if next.Connection == model.ConnectionJump {
					continue // 后继动作显式跳跃，不视为断裂
				}
				gap := next.StartBeat - cur.EndBeat - 1
				if gap > 0 {
					out.BrokenCount++
					out.Broken = append(out.Broken, model.BrokenSegment{
						MovementID: cur.ID,
						DancerNo:   cur.DancerNo,
						ActionName: cur.ActionName,
						EndBeat:    cur.EndBeat,
						NextAction: next.ActionName,
						NextBeat:   next.StartBeat,
						GapBeats:   gap,
					})
				}
			}
		}
	}

	// 稳定输出顺序（按舞者、节拍）。
	sort.Slice(out.Broken, func(i, j int) bool {
		if out.Broken[i].DancerNo == out.Broken[j].DancerNo {
			return out.Broken[i].EndBeat < out.Broken[j].EndBeat
		}
		return out.Broken[i].DancerNo < out.Broken[j].DancerNo
	})
	return out
}

// HasBreak 是否存在动作断裂。
func HasBreak(c *model.ContinuityCheck) bool {
	return c.BrokenCount > 0
}
