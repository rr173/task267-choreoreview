// 民俗舞蹈动作谱记复核台领域实体与错误定义。
package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// 舞段项目状态机：整理中 → 待对齐 → 待复核 → 已发布 → 封存。
// 封存为终态，拒绝任何修改。
const (
	DanceStatusOrganizing = "organizing"
	DanceStatusAligning   = "aligning"
	DanceStatusReviewing  = "reviewing"
	DanceStatusPublished  = "published"
	DanceStatusSealed     = "sealed"
)

// 动作单元状态：候选 → 已对齐 / 连接中断 → 排除。
const (
	MovementStatusCandidate   = "candidate"
	MovementStatusAligned     = "aligned"
	MovementStatusBroken      = "broken"
	MovementStatusExcluded    = "excluded"
)

// 队形关系状态：原始 → 连续 / 冲突 → 确认 / 否决。
const (
	FormationStatusRaw       = "raw"
	FormationStatusContinuous = "continuous"
	FormationStatusConflict  = "conflict"
	FormationStatusConfirmed = "confirmed"
	FormationStatusRejected  = "rejected"
)

// 谱记版本状态：草稿 → 共享 → 冻结 → 替代。
const (
	VersionStatusDraft      = "draft"
	VersionStatusShared     = "shared"
	VersionStatusFrozen     = "frozen"
	VersionStatusSuperseded = "superseded"
)

// 异读候选类别。
const (
	VariantTypeMovement = "movement"
	VariantTypeFormation = "formation"
	VariantTypeBeat     = "beat"
)

// 异读候选状态：候选 → 确认 / 否决（地方变体归入确认）。
const (
	VariantStatusCandidate = "candidate"
	VariantStatusConfirmed = "confirmed"
	VariantStatusRejected  = "rejected"
)

// 动作连接类型：连续（前一动作末拍=后一动作首拍）与跳跃。
const (
	ConnectionContinuous = "continuous"
	ConnectionJump       = "jump"
)

// DancePiece 舞段项目：一段民俗舞蹈谱记复核的容器。
type DancePiece struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Region     string    `json:"region"`
	Dancers    int       `json:"dancers"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MovementUnit 动作单元：舞者执行的一个谱记动作，含起止节拍与连接类型。
type MovementUnit struct {
	ID           int64     `json:"id"`
	DanceID      int64     `json:"dance_id"`
	DancerNo     int       `json:"dancer_no"`
	ActionName   string    `json:"action_name"`
	StartBeat    int       `json:"start_beat"`
	EndBeat      int       `json:"end_beat"`
	Connection   string    `json:"connection"`
	Status       string    `json:"status"`
	Fingerprint  string    `json:"fingerprint"`
	CreatedAt    time.Time `json:"created_at"`
}

// BeatAnchor 节拍锚点：影像证据中检出的节奏点（谱记节拍号 → 影像时间）。
type BeatAnchor struct {
	ID        int64     `json:"id"`
	DanceID   int64     `json:"dance_id"`
	BeatNo    int       `json:"beat_no"`
	ImageTime float64   `json:"image_time"` // 影像时间线秒数
	Confidence float64  `json:"confidence"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// FormationRelation 队形关系：两名舞者间的空间拓扑边。
type FormationRelation struct {
	ID          int64     `json:"id"`
	DanceID     int64     `json:"dance_id"`
	FromDancer  int       `json:"from_dancer"`
	ToDancer    int       `json:"to_dancer"`
	Relation    string    `json:"relation"` // 并肩/斜线/环绕/独立
	BeatStart   int       `json:"beat_start"`
	BeatEnd     int       `json:"beat_end"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// VariantCandidate 异读候选：谱记与影像证据不一致处的复核对象。
type VariantCandidate struct {
	ID         int64     `json:"id"`
	DanceID    int64     `json:"dance_id"`
	Type       string    `json:"type"`
	RefID      int64     `json:"ref_id"`   // 关联动作/队形/节拍 ID
	Detail     string    `json:"detail"`
	Status     string    `json:"status"`
	Verdict    string    `json:"verdict"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
	DecidedAt  *time.Time `json:"decided_at,omitempty"`
}

// NotationVersion 谱记版本：可引用的动作证据包。
type NotationVersion struct {
	ID        int64     `json:"id"`
	DanceID   int64     `json:"dance_id"`
	VersionNo int       `json:"version_no"`
	Status    string    `json:"status"`
	Summary   string    `json:"summary"`
	Evidence  string    `json:"evidence"`
	CreatedAt time.Time `json:"created_at"`
}

// 校验错误集合。
var (
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrInvalidState      = errors.New("invalid state transition")
	ErrSealed            = errors.New("dance is sealed")
	ErrDuplicate         = errors.New("duplicate fingerprint")
	ErrBadRequest        = errors.New("bad request")
	ErrVersionConflict   = errors.New("version already frozen")
	ErrFormationConflict = errors.New("formation relation conflict")
)

// ValidateDance 校验舞段输入。
func ValidateDance(name, region string, dancers int) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: dance name required", ErrBadRequest)
	}
	if dancers < 1 {
		return fmt.Errorf("%w: dancers must be >= 1", ErrBadRequest)
	}
	return nil
}

// ValidateMovement 校验动作单元输入。
func ValidateMovement(dancerNo, startBeat, endBeat int, actionName string) error {
	if strings.TrimSpace(actionName) == "" {
		return fmt.Errorf("%w: action name required", ErrBadRequest)
	}
	if dancerNo < 1 {
		return fmt.Errorf("%w: dancer number must be >= 1", ErrBadRequest)
	}
	if startBeat < 1 || endBeat < startBeat {
		return fmt.Errorf("%w: beat range [%d,%d] invalid", ErrBadRequest, startBeat, endBeat)
	}
	return nil
}

// ValidateBeatAnchor 校验节拍锚点输入。
func ValidateBeatAnchor(beatNo int, imageTime, confidence float64) error {
	if beatNo < 1 {
		return fmt.Errorf("%w: beat number must be >= 1", ErrBadRequest)
	}
	if imageTime < 0 {
		return fmt.Errorf("%w: image time cannot be negative", ErrBadRequest)
	}
	if confidence < 0 || confidence > 1 {
		return fmt.Errorf("%w: confidence must be in [0,1]", ErrBadRequest)
	}
	return nil
}

// ValidateFormation 校验队形关系输入。
func ValidateFormation(fromDancer, toDancer, beatStart, beatEnd int, relation string) error {
	if fromDancer < 1 || toDancer < 1 {
		return fmt.Errorf("%w: dancer numbers must be >= 1", ErrBadRequest)
	}
	if fromDancer == toDancer {
		return fmt.Errorf("%w: formation cannot be self-loop", ErrBadRequest)
	}
	if beatStart < 1 || beatEnd < beatStart {
		return fmt.Errorf("%w: beat range invalid", ErrBadRequest)
	}
	return nil
}

// CanTransitDance 舞段状态机：合法的状态转移表。
func CanTransitDance(from, to string) bool {
	switch from {
	case DanceStatusOrganizing:
		return to == DanceStatusAligning
	case DanceStatusAligning:
		return to == DanceStatusReviewing
	case DanceStatusReviewing:
		return to == DanceStatusPublished
	case DanceStatusPublished:
		return to == DanceStatusSealed
	}
	return false
}

// CanTransitVersion 谱记版本状态机。
func CanTransitVersion(from, to string) bool {
	switch from {
	case VersionStatusDraft:
		return to == VersionStatusShared
	case VersionStatusShared:
		return to == VersionStatusFrozen
	case VersionStatusFrozen:
		return to == VersionStatusSuperseded
	}
	return false
}
