// 队形模块：舞者间队形拓扑校验。
package formation

import (
	"sort"
	"time"

	"task267-choreoreview/internal/model"
)

// Service 队形拓扑校验服务。
type Service struct{}

// New 构造队形服务。
func New() *Service {
	return &Service{}
}

// Check 校验队形拓扑：
//   - 同一对舞者 (from,to) 在同一节拍区间内不得同时存在两条互斥关系；
//   - 关系值集合：并肩(line)、斜线(diagonal)、环绕(encircle)、独立(apart)；
//   - 方向性：并肩/斜线视为对称，环绕/独立视为无向，检查双向冲突。
func (s *Service) Check(edges []*model.FormationRelation) *model.FormationCheck {
	out := &model.FormationCheck{
		TotalEdges: len(edges),
		CheckedAt:  time.Now().UTC(),
	}

	// 关键：同一对舞者、时间重叠、关系不同 → 冲突。
	// 建立规范化键：min, max 排序，忽略方向（对对称关系）。
	normalize := func(a, b int) (int, int) {
		if a < b {
			return a, b
		}
		return b, a
	}

	for i := 0; i < len(edges); i++ {
		ei := edges[i]
		if ei.Status == model.FormationStatusRejected {
			continue
		}
		for j := i + 1; j < len(edges); j++ {
			ej := edges[j]
			if ej.Status == model.FormationStatusRejected {
				continue
			}
			a1, b1 := normalize(ei.FromDancer, ei.ToDancer)
			a2, b2 := normalize(ej.FromDancer, ej.ToDancer)
			if a1 != a2 || b1 != b2 {
				continue
			}
			if !overlap(ei.BeatStart, ei.BeatEnd, ej.BeatStart, ej.BeatEnd) {
				continue
			}
			if ei.Relation == ej.Relation {
				// 同关系重复（同一区间重复录入）
				out.ConflictCount++
				out.Conflicts = append(out.Conflicts, model.ConflictEdge{
					RelationID: ej.ID,
					FromDancer: ej.FromDancer,
					ToDancer:   ej.ToDancer,
					Relation:   ej.Relation,
					BeatStart:  ej.BeatStart,
					BeatEnd:    ej.BeatEnd,
					Reason:     "duplicate relation for same pair",
				})
				continue
			}
			// 不同关系：并肩 vs 斜线 或 环绕 vs 独立 均冲突
			out.ConflictCount++
			out.Conflicts = append(out.Conflicts, model.ConflictEdge{
				RelationID: ej.ID,
				FromDancer: ej.FromDancer,
				ToDancer:   ej.ToDancer,
				Relation:   ej.Relation,
				BeatStart:  ej.BeatStart,
				BeatEnd:    ej.BeatEnd,
				Reason:     "mutually exclusive relation vs " + ei.Relation,
			})
		}
	}

	sort.Slice(out.Conflicts, func(i, j int) bool {
		return out.Conflicts[i].RelationID < out.Conflicts[j].RelationID
	})
	return out
}

// overlap 两个闭区间 [a,b] 与 [c,d] 是否重叠。
func overlap(aStart, aEnd, bStart, bEnd int) bool {
	return aStart <= bEnd && bStart <= aEnd
}
