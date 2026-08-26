// 舞段项目仓储。
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"task267-choreoreview/internal/model"
)

// DanceStore 管理舞段项目。
type DanceStore struct {
	db *sql.DB
	mu *sync.Mutex
}

// Create 创建舞段（名称唯一）。
func (s *DanceStore) Create(name, region string, dancers int, notes string) (*model.DancePiece, error) {
	now := time.Now().UTC()
	res, err := s.db.Exec(
		`INSERT INTO dances(name, region, dancers, status, notes, created_at, updated_at)
		 VALUES(?,?,?,?,?,?,?)`,
		name, region, dancers, model.DanceStatusOrganizing, notes, now, now,
	)
	if err != nil {
		return nil, mapUniqueErr(err, "dance name already exists")
	}
	id, _ := res.LastInsertId()
	return s.Get(id)
}

// Get 按 ID 查询舞段。
func (s *DanceStore) Get(id int64) (*model.DancePiece, error) {
	row := s.db.QueryRow(
		`SELECT id, name, region, dancers, status, notes, created_at, updated_at FROM dances WHERE id=?`, id)
	var d model.DancePiece
	if err := row.Scan(&d.ID, &d.Name, &d.Region, &d.Dancers, &d.Status, &d.Notes, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: dance %d", model.ErrNotFound, id)
		}
		return nil, err
	}
	return &d, nil
}

// List 列出全部舞段。
func (s *DanceStore) List() ([]*model.DancePiece, error) {
	rows, err := s.db.Query(
		`SELECT id, name, region, dancers, status, notes, created_at, updated_at FROM dances ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.DancePiece
	for rows.Next() {
		var d model.DancePiece
		if err := rows.Scan(&d.ID, &d.Name, &d.Region, &d.Dancers, &d.Status, &d.Notes, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &d)
	}
	return out, rows.Err()
}

// SetStatus 状态流转（带状态机校验）。
func (s *DanceStore) SetStatus(id int64, to string) (*model.DancePiece, error) {
	d, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if d.Status == model.DanceStatusSealed {
		return nil, fmt.Errorf("%w: dance %d", model.ErrSealed, id)
	}
	now := time.Now().UTC()
	if _, err := s.db.Exec(`UPDATE dances SET status=?, updated_at=? WHERE id=?`, to, now, id); err != nil {
		return nil, err
	}
	return s.Get(id)
}

// Seal 封存舞段（终态，拒绝后续修改）。
func (s *DanceStore) Seal(id int64) (*model.DancePiece, error) {
	return s.SetStatus(id, model.DanceStatusSealed)
}

// EnsureMutable 校验舞段未封存；已封存返回 ErrSealed。
func (s *DanceStore) EnsureMutable(id int64) error {
	d, err := s.Get(id)
	if err != nil {
		return err
	}
	if d.Status == model.DanceStatusSealed {
		return fmt.Errorf("%w: dance %d", model.ErrSealed, id)
	}
	return nil
}
