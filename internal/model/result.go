// 业务结果类型：分析结论与统计摘要。
package model

import "time"

// AlignmentResult 节拍对齐结果：谱记节拍与影像锚点的偏移分析。
type AlignmentResult struct {
	BeatCount    int             `json:"beat_count"`
	AnchorCount  int             `json:"anchor_count"`
	MatchedCount int             `json:"matched_count"`
	MaxOffsetMs  float64         `json:"max_offset_ms"`
	AvgOffsetMs  float64         `json:"avg_offset_ms"`
	Offsets      []BeatOffset    `json:"offsets"`
	GeneratedAt  time.Time       `json:"generated_at"`
}

// BeatOffset 单个节拍的偏移。
type BeatOffset struct {
	BeatNo     int     `json:"beat_no"`
	ImageTime  float64 `json:"image_time"`
	OffsetMs   float64 `json:"offset_ms"`
	Anchored   bool    `json:"anchored"`
}

// ContinuityCheck 动作连续性校验结果。
type ContinuityCheck struct {
	TotalUnits   int               `json:"total_units"`
	BrokenCount  int               `json:"broken_count"`
	Broken       []BrokenSegment   `json:"broken"`
	CheckedAt    time.Time         `json:"checked_at"`
}

// BrokenSegment 断裂段：动作连接跳跃处。
type BrokenSegment struct {
	MovementID int64  `json:"movement_id"`
	DancerNo   int    `json:"dancer_no"`
	ActionName string `json:"action_name"`
	EndBeat    int    `json:"end_beat"`
	NextAction string `json:"next_action,omitempty"`
	NextBeat   int    `json:"next_beat,omitempty"`
	GapBeats   int    `json:"gap_beats"`
}

// FormationCheck 队形拓扑校验结果。
type FormationCheck struct {
	TotalEdges     int            `json:"total_edges"`
	ConflictCount  int            `json:"conflict_count"`
	Conflicts      []ConflictEdge `json:"conflicts"`
	CheckedAt      time.Time      `json:"checked_at"`
}

// ConflictEdge 队形冲突边：同一对舞者同时存在互斥关系。
type ConflictEdge struct {
	RelationID int64  `json:"relation_id"`
	FromDancer int    `json:"from_dancer"`
	ToDancer   int    `json:"to_dancer"`
	Relation   string `json:"relation"`
	BeatStart  int    `json:"beat_start"`
	BeatEnd    int    `json:"beat_end"`
	Reason     string `json:"reason"`
}

// Stats 统计摘要。
type Stats struct {
	DanceCount       int `json:"dance_count"`
	MovementCount    int `json:"movement_count"`
	BeatAnchorCount  int `json:"beat_anchor_count"`
	FormationCount   int `json:"formation_count"`
	VariantCount     int `json:"variant_count"`
	OpenVariantCount int `json:"open_variant_count"`
	VersionCount     int `json:"version_count"`
}
