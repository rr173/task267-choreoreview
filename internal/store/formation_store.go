// 队形关系仓储。
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"task267-choreoreview/internal/model"
)

// FormationStore 管理舞者间队形拓扑边。
type FormationStore struct {
	db *sql.DB
}

// Create 导入队形关系；同一舞段内 (from_dancer, to_dancer, beat_start) 唯一。
func (s *FormationStore) Create(f *model.FormationRelation) (*model.FormationRelation, error) {
	now := time.Now().UTC()
	res, err := s.db.Exec(
		`INSERT INTO formations(dance_id, from_dancer, to_dancer, relation, beat_start, beat_end, status, created_at)
		 VALUES(?,?,?,?,?,?,?,?)`,
		f.DanceID, f.FromDancer, f.ToDancer, f.Relation, f.BeatStart, f.BeatEnd, f.Status, now,
	)
	if err != nil {
		return nil, mapUniqueErr(err, "formation duplicate")
	}
	id, _ := res.LastInsertId()
	return s.Get(id)
}

// Get 按 ID 查询。
func (s *FormationStore) Get(id int64) (*model.FormationRelation, error) {
	row := s.db.QueryRow(
		`SELECT id, dance_id, from_dancer, to_dancer, relation, beat_start, beat_end, status, created_at
		 FROM formations WHERE id=?`, id)
	var f model.FormationRelation
	if err := row.Scan(&f.ID, &f.DanceID, &f.FromDancer, &f.ToDancer, &f.Relation,
		&f.BeatStart, &f.BeatEnd, &f.Status, &f.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: formation %d", model.ErrNotFound, id)
		}
		return nil, err
	}
	return &f, nil
}

// ListByDance 列出舞段全部队形边。
func (s *FormationStore) ListByDance(danceID int64) ([]*model.FormationRelation, error) {
	rows, err := s.db.Query(
		`SELECT id, dance_id, from_dancer, to_dancer, relation, beat_start, beat_end, status, created_at
		 FROM formations WHERE dance_id=? ORDER BY from_dancer, to_dancer, beat_start`, danceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.FormationRelation
	for rows.Next() {
		var f model.FormationRelation
		if err := rows.Scan(&f.ID, &f.DanceID, &f.FromDancer, &f.ToDancer, &f.Relation,
			&f.BeatStart, &f.BeatEnd, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &f)
	}
	return out, rows.Err()
}

// SetStatus 更新队形边状态。
func (s *FormationStore) SetStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE formations SET status=? WHERE id=?`, status, id)
	return err
}

// CountByDance 统计舞段队形边数。
func (s *FormationStore) CountByDance(danceID int64) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM formations WHERE dance_id=?`, danceID).Scan(&n)
	return n, err
}
