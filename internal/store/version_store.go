// 谱记版本仓储。
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"task267-choreoreview/internal/model"
)

// VersionStore 管理可引用的谱记版本。
type VersionStore struct {
	db *sql.DB
}

// NextVersionNo 计算舞段下一个版本号。
func (s *VersionStore) NextVersionNo(danceID int64) (int, error) {
	var max int
	err := s.db.QueryRow(
		`SELECT COALESCE(MAX(version_no),0) FROM notation_versions WHERE dance_id=?`, danceID).Scan(&max)
	return max + 1, err
}

// Create 创建谱记版本（草稿）。
func (s *VersionStore) Create(danceID int64, summary, evidence string) (*model.NotationVersion, error) {
	no, err := s.NextVersionNo(danceID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	res, err := s.db.Exec(
		`INSERT INTO notation_versions(dance_id, version_no, status, summary, evidence, created_at)
		 VALUES(?,?,?,?,?,?)`,
		danceID, no, model.VersionStatusDraft, summary, evidence, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.Get(id)
}

// Get 按 ID 查询。
func (s *VersionStore) Get(id int64) (*model.NotationVersion, error) {
	row := s.db.QueryRow(
		`SELECT id, dance_id, version_no, status, summary, evidence, created_at FROM notation_versions WHERE id=?`, id)
	var v model.NotationVersion
	if err := row.Scan(&v.ID, &v.DanceID, &v.VersionNo, &v.Status, &v.Summary, &v.Evidence, &v.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: version %d", model.ErrNotFound, id)
		}
		return nil, err
	}
	return &v, nil
}

// ListByDance 列出舞段全部版本。
func (s *VersionStore) ListByDance(danceID int64) ([]*model.NotationVersion, error) {
	rows, err := s.db.Query(
		`SELECT id, dance_id, version_no, status, summary, evidence, created_at
		 FROM notation_versions WHERE dance_id=? ORDER BY version_no`, danceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.NotationVersion
	for rows.Next() {
		var v model.NotationVersion
		if err := rows.Scan(&v.ID, &v.DanceID, &v.VersionNo, &v.Status, &v.Summary, &v.Evidence, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}

// SetStatus 状态流转（草稿→共享→冻结→替代）。
func (s *VersionStore) SetStatus(id int64, to string) (*model.NotationVersion, error) {
	v, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitVersion(v.Status, to) {
		return nil, fmt.Errorf("%w: version %d %s -> %s", model.ErrInvalidState, id, v.Status, to)
	}
	if _, err := s.db.Exec(`UPDATE notation_versions SET status=? WHERE id=?`, to, id); err != nil {
		return nil, err
	}
	return s.Get(id)
}

// Freeze 冻结版本（不可再修改）。必须经由 shared 状态流转，不得从 draft 直接冻结。
func (s *VersionStore) Freeze(id int64) (*model.NotationVersion, error) {
	return s.SetStatus(id, model.VersionStatusFrozen)
}
