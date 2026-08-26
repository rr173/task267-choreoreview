// 动作单元仓储。
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"task267-choreoreview/internal/model"
)

// MovementStore 管理舞段内的动作单元。
type MovementStore struct {
	db *sql.DB
}

// Create 导入动作单元；同一舞段内 (dancer_no, start_beat) 唯一。
func (s *MovementStore) Create(m *model.MovementUnit) (*model.MovementUnit, error) {
	now := time.Now().UTC()
	fp := fmt.Sprintf("mv:%d:%d:%d:%s", m.DanceID, m.DancerNo, m.StartBeat, m.ActionName)
	res, err := s.db.Exec(
		`INSERT INTO movements(dance_id, dancer_no, action_name, start_beat, end_beat, connection, status, fingerprint, created_at)
		 VALUES(?,?,?,?,?,?,?,?,?)`,
		m.DanceID, m.DancerNo, m.ActionName, m.StartBeat, m.EndBeat, m.Connection, m.Status, fp, now,
	)
	if err != nil {
		return nil, mapUniqueErr(err, "movement duplicate (dancer, start_beat)")
	}
	id, _ := res.LastInsertId()
	return s.Get(id)
}

// Get 按 ID 查询。
func (s *MovementStore) Get(id int64) (*model.MovementUnit, error) {
	row := s.db.QueryRow(
		`SELECT id, dance_id, dancer_no, action_name, start_beat, end_beat, connection, status, fingerprint, created_at
		 FROM movements WHERE id=?`, id)
	var m model.MovementUnit
	if err := row.Scan(&m.ID, &m.DanceID, &m.DancerNo, &m.ActionName, &m.StartBeat, &m.EndBeat,
		&m.Connection, &m.Status, &m.Fingerprint, &m.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: movement %d", model.ErrNotFound, id)
		}
		return nil, err
	}
	return &m, nil
}

// ListByDance 列出舞段全部动作单元（按舞者、节拍排序）。
func (s *MovementStore) ListByDance(danceID int64) ([]*model.MovementUnit, error) {
	rows, err := s.db.Query(
		`SELECT id, dance_id, dancer_no, action_name, start_beat, end_beat, connection, status, fingerprint, created_at
		 FROM movements WHERE dance_id=? ORDER BY dancer_no, start_beat`, danceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.MovementUnit
	for rows.Next() {
		var m model.MovementUnit
		if err := rows.Scan(&m.ID, &m.DanceID, &m.DancerNo, &m.ActionName, &m.StartBeat, &m.EndBeat,
			&m.Connection, &m.Status, &m.Fingerprint, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

// CountByDance 统计舞段动作单元数。
func (s *MovementStore) CountByDance(danceID int64) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM movements WHERE dance_id=?`, danceID).Scan(&n)
	return n, err
}

// SetStatus 更新动作单元状态。
func (s *MovementStore) SetStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE movements SET status=? WHERE id=?`, status, id)
	return err
}

// MarkBroken 将动作单元标为连接中断。
func (s *MovementStore) MarkBroken(id int64) error {
	return s.SetStatus(id, model.MovementStatusBroken)
}

// MarkAligned 将动作单元标为已对齐。
func (s *MovementStore) MarkAligned(id int64) error {
	return s.SetStatus(id, model.MovementStatusAligned)
}
