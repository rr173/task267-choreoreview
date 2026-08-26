// 服务编排层：串联仓储与业务包，承载端到端用例。
package service

import (
	"errors"
	"fmt"

	"task267-choreoreview/internal/beat"
	"task267-choreoreview/internal/formation"
	"task267-choreoreview/internal/model"
	"task267-choreoreview/internal/notation"
	"task267-choreoreview/internal/store"
	"task267-choreoreview/internal/variant"
)

// Service 聚合全部编排动作。
type Service struct {
	Store *store.Store
}

// New 构造服务编排。
func New(st *store.Store) *Service {
	return &Service{Store: st}
}

// Dance 舞段编排组。
type DanceActions struct{ svc *Service }

// Dance 返回舞段动作。
func (s *Service) Dance() *DanceActions { return &DanceActions{svc: s} }

// CreateDance 创建舞段项目。
func (d *DanceActions) CreateDance(name, region string, dancers int, notes string) (*model.DancePiece, error) {
	if err := model.ValidateDance(name, region, dancers); err != nil {
		return nil, err
	}
	return d.svc.Store.Dances.Create(name, region, dancers, notes)
}

// ListDances 列出舞段。
func (d *DanceActions) ListDances() ([]*model.DancePiece, error) {
	return d.svc.Store.Dances.List()
}

// GetDance 查询舞段。
func (d *DanceActions) GetDance(id int64) (*model.DancePiece, error) {
	return d.svc.Store.Dances.Get(id)
}

// TransitDance 舞段状态流转。
func (d *DanceActions) TransitDance(id int64, to string) (*model.DancePiece, error) {
	return d.svc.Store.Dances.SetStatus(id, to)
}

// SealDance 封存舞段。
func (d *DanceActions) SealDance(id int64) (*model.DancePiece, error) {
	return d.svc.Store.Dances.Seal(id)
}

// Movement 动作编排组。
type MovementActions struct{ svc *Service }

// Movement 返回动作动作。
func (s *Service) Movement() *MovementActions { return &MovementActions{svc: s} }

// AddMovement 导入动作单元。
func (m *MovementActions) AddMovement(danceID int64, dancerNo, startBeat, endBeat int, actionName, connection string) (*model.MovementUnit, error) {
	if err := model.ValidateMovement(dancerNo, startBeat, endBeat, actionName); err != nil {
		return nil, err
	}
	if connection == "" {
		connection = model.ConnectionContinuous
	}
	if connection != model.ConnectionContinuous && connection != model.ConnectionJump {
		return nil, fmt.Errorf("%w: connection must be continuous or jump", model.ErrBadRequest)
	}
	if err := m.svc.Store.Dances.EnsureMutable(danceID); err != nil {
		return nil, err
	}
	unit := &model.MovementUnit{
		DanceID:    danceID,
		DancerNo:   dancerNo,
		ActionName: actionName,
		StartBeat:  startBeat,
		EndBeat:    endBeat,
		Connection: connection,
		Status:     model.MovementStatusCandidate,
	}
	return m.svc.Store.Movements.Create(unit)
}

// ListMovements 列出舞段动作。
func (m *MovementActions) ListMovements(danceID int64) ([]*model.MovementUnit, error) {
	if _, err := m.svc.Store.Dances.Get(danceID); err != nil {
		return nil, err
	}
	return m.svc.Store.Movements.ListByDance(danceID)
}

// Beat 节拍编排组。
type BeatActions struct{ svc *Service }

// Beat 返回节拍动作。
func (s *Service) Beat() *BeatActions { return &BeatActions{svc: s} }

// AddBeatAnchor 导入节拍锚点。
func (b *BeatActions) AddBeatAnchor(danceID int64, beatNo int, imageTime, confidence float64, source string) (*model.BeatAnchor, error) {
	if err := model.ValidateBeatAnchor(beatNo, imageTime, confidence); err != nil {
		return nil, err
	}
	if err := b.svc.Store.Dances.EnsureMutable(danceID); err != nil {
		return nil, err
	}
	anchor := &model.BeatAnchor{
		DanceID:    danceID,
		BeatNo:     beatNo,
		ImageTime:  imageTime,
		Confidence: confidence,
		Source:     source,
	}
	return b.svc.Store.Beats.Create(anchor)
}

// ListBeatAnchors 列出节拍锚点。
func (b *BeatActions) ListBeatAnchors(danceID int64) ([]*model.BeatAnchor, error) {
	if _, err := b.svc.Store.Dances.Get(danceID); err != nil {
		return nil, err
	}
	return b.svc.Store.Beats.ListByDance(danceID)
}

// Formation 队形编排组。
type FormationActions struct{ svc *Service }

// Formation 返回队形动作。
func (s *Service) Formation() *FormationActions { return &FormationActions{svc: s} }

// AddFormation 导入队形关系。
func (f *FormationActions) AddFormation(danceID int64, fromDancer, toDancer, beatStart, beatEnd int, relation string) (*model.FormationRelation, error) {
	if err := model.ValidateFormation(fromDancer, toDancer, beatStart, beatEnd, relation); err != nil {
		return nil, err
	}
	if relation != "line" && relation != "diagonal" && relation != "encircle" && relation != "apart" {
		return nil, fmt.Errorf("%w: relation must be line/diagonal/encircle/apart", model.ErrBadRequest)
	}
	if err := f.svc.Store.Dances.EnsureMutable(danceID); err != nil {
		return nil, err
	}
	edge := &model.FormationRelation{
		DanceID:    danceID,
		FromDancer: fromDancer,
		ToDancer:   toDancer,
		Relation:   relation,
		BeatStart:  beatStart,
		BeatEnd:    beatEnd,
		Status:     model.FormationStatusRaw,
	}
	return f.svc.Store.Formations.Create(edge)
}

// ListFormations 列出队形关系。
func (f *FormationActions) ListFormations(danceID int64) ([]*model.FormationRelation, error) {
	if _, err := f.svc.Store.Dances.Get(danceID); err != nil {
		return nil, err
	}
	return f.svc.Store.Formations.ListByDance(danceID)
}

// Analyze 分析编排组：对齐、连续性与拓扑校验、异读物化。
type AnalyzeActions struct{ svc *Service }

// Analyze 返回分析动作。
func (s *Service) Analyze() *AnalyzeActions { return &AnalyzeActions{svc: s} }

// AlignBeats 节拍对齐分析。
func (a *AnalyzeActions) AlignBeats(danceID int64) (*model.AlignmentResult, error) {
	anchors, err := a.svc.Store.Beats.ListByDance(danceID)
	if err != nil {
		return nil, err
	}
	return beat.New().Align(anchors), nil
}

// CheckContinuity 动作连续性校验。
func (a *AnalyzeActions) CheckContinuity(danceID int64) (*model.ContinuityCheck, error) {
	units, err := a.svc.Store.Movements.ListByDance(danceID)
	if err != nil {
		return nil, err
	}
	return notation.New().CheckContinuity(units), nil
}

// CheckFormation 队形拓扑校验。
func (a *AnalyzeActions) CheckFormation(danceID int64) (*model.FormationCheck, error) {
	edges, err := a.svc.Store.Formations.ListByDance(danceID)
	if err != nil {
		return nil, err
	}
	return formation.New().Check(edges), nil
}

// RunFullAnalysis 全量分析：物化全部异读候选（幂等，重复运行不产生重复候选）。
func (a *AnalyzeActions) RunFullAnalysis(danceID int64) (*FullAnalysisResult, error) {
	if _, err := a.svc.Store.Dances.Get(danceID); err != nil {
		return nil, err
	}
	res := &FullAnalysisResult{DanceID: danceID}

	// 1) 动作连续性
	cont, err := a.CheckContinuity(danceID)
	if err != nil {
		return nil, err
	}
	res.Continuity = cont
	for _, c := range variant.New().BuildMovementCandidates(danceID, cont) {
		if _, err := a.svc.Store.Variants.Create(c); err != nil {
			// 幂等：重复候选直接跳过
			if !isDuplicate(err) {
				return nil, err
			}
		}
	}

	// 2) 节拍对齐
	align, err := a.AlignBeats(danceID)
	if err != nil {
		return nil, err
	}
	res.Alignment = align
	for _, c := range variant.New().BuildBeatCandidates(danceID, align, 250.0) {
		if _, err := a.svc.Store.Variants.Create(c); err != nil {
			if !isDuplicate(err) {
				return nil, err
			}
		}
	}

	// 3) 队形拓扑
	form, err := a.CheckFormation(danceID)
	if err != nil {
		return nil, err
	}
	res.Formation = form
	for _, c := range variant.New().BuildFormationCandidates(danceID, form) {
		if _, err := a.svc.Store.Variants.Create(c); err != nil {
			if !isDuplicate(err) {
				return nil, err
			}
		}
	}

	// 4) 标记断裂动作为 broken
	if cont.BrokenCount > 0 {
		for _, b := range cont.Broken {
			_ = a.svc.Store.Movements.MarkBroken(b.MovementID)
		}
	}

	open, err := a.svc.Store.Variants.OpenByDance(danceID)
	if err != nil {
		return nil, err
	}
	res.OpenCandidates = open
	return res, nil
}

// FullAnalysisResult 全量分析结果。
type FullAnalysisResult struct {
	DanceID        int64                      `json:"dance_id"`
	Continuity     *model.ContinuityCheck     `json:"continuity"`
	Alignment      *model.AlignmentResult     `json:"alignment"`
	Formation      *model.FormationCheck      `json:"formation"`
	OpenCandidates []*model.VariantCandidate  `json:"open_candidates"`
}

// Variant 裁决编排组。
type VariantActions struct{ svc *Service }

// Variant 返回裁决动作。
func (s *Service) Variant() *VariantActions { return &VariantActions{svc: s} }

// ListVariants 列出舞段全部候选。
func (v *VariantActions) ListVariants(danceID int64) ([]*model.VariantCandidate, error) {
	if _, err := v.svc.Store.Dances.Get(danceID); err != nil {
		return nil, err
	}
	return v.svc.Store.Variants.ListByDance(danceID)
}

// OpenVariants 列出未裁决候选。
func (v *VariantActions) OpenVariants(danceID int64) ([]*model.VariantCandidate, error) {
	if _, err := v.svc.Store.Dances.Get(danceID); err != nil {
		return nil, err
	}
	return v.svc.Store.Variants.OpenByDance(danceID)
}

// Adjudicate 裁决异读（确认/保留地方变体/否决）。
func (v *VariantActions) Adjudicate(variantID int64, verdict, reason string) (*model.VariantCandidate, error) {
	c, err := v.svc.Store.Variants.Get(variantID)
	if err != nil {
		return nil, err
	}
	dance, err := v.svc.Store.Dances.Get(c.DanceID)
	if err != nil {
		return nil, err
	}
	if dance.Status == model.DanceStatusSealed {
		return nil, fmt.Errorf("%w: dance %d", model.ErrSealed, dance.ID)
	}
	// 队形候选确认时同步更新队形边状态。
	if c.Type == model.VariantTypeFormation && verdict == model.VariantStatusConfirmed {
		_ = v.svc.Store.Formations.SetStatus(c.RefID, model.FormationStatusConfirmed)
	}
	if c.Type == model.VariantTypeFormation && verdict == model.VariantStatusRejected {
		_ = v.svc.Store.Formations.SetStatus(c.RefID, model.FormationStatusRejected)
	}
	return v.svc.Store.Variants.Adjudicate(variantID, verdict, reason)
}

// Version 版本编排组。
type VersionActions struct{ svc *Service }

// Version 返回版本动作。
func (s *Service) Version() *VersionActions { return &VersionActions{svc: s} }

// CreateVersion 创建谱记版本草稿（仅允许在已发布/待复核状态）。
func (v *VersionActions) CreateVersion(danceID int64, summary, evidence string) (*model.NotationVersion, error) {
	dance, err := v.svc.Store.Dances.Get(danceID)
	if err != nil {
		return nil, err
	}
	if dance.Status != model.DanceStatusReviewing && dance.Status != model.DanceStatusPublished {
		return nil, fmt.Errorf("%w: dance must be reviewing or published, got %s", model.ErrInvalidState, dance.Status)
	}
	return v.svc.Store.Versions.Create(danceID, summary, evidence)
}

// ListVersions 列出舞段版本。
func (v *VersionActions) ListVersions(danceID int64) ([]*model.NotationVersion, error) {
	if _, err := v.svc.Store.Dances.Get(danceID); err != nil {
		return nil, err
	}
	return v.svc.Store.Versions.ListByDance(danceID)
}

// FreezeVersion 冻结版本。
func (v *VersionActions) FreezeVersion(id int64) (*model.NotationVersion, error) {
	return v.svc.Store.Versions.Freeze(id)
}

// SupersedeVersion 替代版本（冻结 → 替代）。
func (v *VersionActions) SupersedeVersion(id int64) (*model.NotationVersion, error) {
	return v.svc.Store.Versions.SetStatus(id, model.VersionStatusSuperseded)
}

// Stats 统计摘要。
func (s *Service) Stats() (*model.Stats, error) {
	st := s.Store
	dances, err := st.Dances.List()
	if err != nil {
		return nil, err
	}
	out := &model.Stats{DanceCount: len(dances)}
	for _, d := range dances {
		nm, err := st.Movements.CountByDance(d.ID)
		if err != nil {
			return nil, err
		}
		nb, err := st.Beats.CountByDance(d.ID)
		if err != nil {
			return nil, err
		}
		nf, err := st.Formations.CountByDance(d.ID)
		if err != nil {
			return nil, err
		}
		nv, err := st.Variants.CountOpenByDance(d.ID)
		if err != nil {
			return nil, err
		}
		nvv, err := st.Versions.NextVersionNo(d.ID)
		if err != nil {
			return nil, err
		}
		out.MovementCount += nm
		out.BeatAnchorCount += nb
		out.FormationCount += nf
		out.OpenVariantCount += nv
		out.VariantCount += nv
		out.VersionCount += nvv - 1
	}
	return out, nil
}

func isDuplicate(err error) bool {
	return errors.Is(err, model.ErrDuplicate)
}
