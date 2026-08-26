// 节拍锚点仓储。
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"task267-choreoreview/internal/model"
)

// BeatStore 管理影像证据检出的节拍锚点。
type BeatStore struct {
	db *sql.DB
}

// Create 导入节拍锚点；同一舞段内 beat_no 唯一（幂等）。
func (s *BeatStore) Create(b *model.BeatAnchor) (*model.BeatAnchor, error) {
	now := time.Now().UTC()
	res, err := s.db.Exec(
		`INSERT INTO beat_anchors(dance_id, beat_no, image_time, confidence, source, created_at)
		 VALUES(?,?,?,?,?,?)`,
		b.DanceID, b.BeatNo, b.ImageTime, b.Confidence, b.Source, now,
	)
	if err != nil {
		return nil, mapUniqueErr(err, "beat anchor duplicate")
	}
	id, _ := res.LastInsertId()
	return s.Get(id)
}

// Get 按 ID 查询。
func (s *BeatStore) Get(id int64) (*model.BeatAnchor, error) {
	row := s.db.QueryRow(
		`SELECT id, dance_id, beat_no, image_time, confidence, source, created_at FROM beat_anchors WHERE id=?`, id)
	var b model.BeatAnchor
	if err := row.Scan(&b.ID, &b.DanceID, &b.BeatNo, &b.ImageTime, &b.Confidence, &b.Source, &b.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: beat anchor %d", model.ErrNotFound, id)
		}
		return nil, err
	}
	return &b, nil
}

// ListByDance 列出舞段全部锚点（按节拍号排序）。
func (s *BeatStore) ListByDance(danceID int64) ([]*model.BeatAnchor, error) {
	rows, err := s.db.Query(
		`SELECT id, dance_id, beat_no, image_time, confidence, source, created_at
		 FROM beat_anchors WHERE dance_id=? ORDER BY beat_no`, danceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.BeatAnchor
	for rows.Next() {
		var b model.BeatAnchor
		if err := rows.Scan(&b.ID, &b.DanceID, &b.BeatNo, &b.ImageTime, &b.Confidence, &b.Source, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &b)
	}
	return out, rows.Err()
}

// CountByDance 统计舞段锚点数。
func (s *BeatStore) CountByDance(danceID int64) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM beat_anchors WHERE dance_id=?`, danceID).Scan(&n)
	return n, err
}
