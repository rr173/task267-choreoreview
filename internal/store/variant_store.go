// 异读候选仓储。
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"task267-choreoreview/internal/model"
)

// VariantStore 管理异读候选与裁决。
type VariantStore struct {
	db *sql.DB
}

// Create 生成异读候选；同一舞段内 (type, ref_id) 唯一。
func (s *VariantStore) Create(v *model.VariantCandidate) (*model.VariantCandidate, error) {
	now := time.Now().UTC()
	res, err := s.db.Exec(
		`INSERT INTO variants(dance_id, type, ref_id, detail, status, verdict, reason, created_at)
		 VALUES(?,?,?,?,?,?,?,?)`,
		v.DanceID, v.Type, v.RefID, v.Detail, v.Status, v.Verdict, v.Reason, now,
	)
	if err != nil {
		return nil, mapUniqueErr(err, "variant duplicate for ref")
	}
	id, _ := res.LastInsertId()
	return s.Get(id)
}

// Get 按 ID 查询。
func (s *VariantStore) Get(id int64) (*model.VariantCandidate, error) {
	row := s.db.QueryRow(
		`SELECT id, dance_id, type, ref_id, detail, status, verdict, reason, created_at, decided_at
		 FROM variants WHERE id=?`, id)
	var v model.VariantCandidate
	var decided sql.NullTime
	if err := row.Scan(&v.ID, &v.DanceID, &v.Type, &v.RefID, &v.Detail, &v.Status,
		&v.Verdict, &v.Reason, &v.CreatedAt, &decided); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: variant %d", model.ErrNotFound, id)
		}
		return nil, err
	}
	if decided.Valid {
		v.DecidedAt = &decided.Time
	}
	return &v, nil
}

// ListByDance 列出舞段全部异读候选。
func (s *VariantStore) ListByDance(danceID int64) ([]*model.VariantCandidate, error) {
	rows, err := s.db.Query(
		`SELECT id, dance_id, type, ref_id, detail, status, verdict, reason, created_at, decided_at
		 FROM variants WHERE dance_id=? ORDER BY id`, danceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.VariantCandidate
	for rows.Next() {
		var v model.VariantCandidate
		var decided sql.NullTime
		if err := rows.Scan(&v.ID, &v.DanceID, &v.Type, &v.RefID, &v.Detail, &v.Status,
			&v.Verdict, &v.Reason, &v.CreatedAt, &decided); err != nil {
			return nil, err
		}
		if decided.Valid {
			v.DecidedAt = &decided.Time
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}

// OpenByDance 列出舞段未裁决候选。
func (s *VariantStore) OpenByDance(danceID int64) ([]*model.VariantCandidate, error) {
	rows, err := s.db.Query(
		`SELECT id, dance_id, type, ref_id, detail, status, verdict, reason, created_at, decided_at
		 FROM variants WHERE dance_id=? AND status=? ORDER BY id`, danceID, model.VariantStatusCandidate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.VariantCandidate
	for rows.Next() {
		var v model.VariantCandidate
		var decided sql.NullTime
		if err := rows.Scan(&v.ID, &v.DanceID, &v.Type, &v.RefID, &v.Detail, &v.Status,
			&v.Verdict, &v.Reason, &v.CreatedAt, &decided); err != nil {
			return nil, err
		}
		if decided.Valid {
			v.DecidedAt = &decided.Time
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}

// Adjudicate 裁决异读：确认（含地方变体）或否决。
func (s *VariantStore) Adjudicate(id int64, verdict, reason string) (*model.VariantCandidate, error) {
	v, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if v.Status != model.VariantStatusCandidate {
		return nil, fmt.Errorf("%w: variant %d already decided", model.ErrInvalidState, id)
	}
	if verdict != model.VariantStatusConfirmed && verdict != model.VariantStatusRejected {
		return nil, fmt.Errorf("%w: verdict must be confirmed or rejected", model.ErrBadRequest)
	}
	now := time.Now().UTC()
	if _, err := s.db.Exec(
		`UPDATE variants SET status=?, verdict=?, reason=?, decided_at=? WHERE id=?`,
		verdict, verdict, reason, now, id); err != nil {
		return nil, err
	}
	return s.Get(id)
}

// CountByDance 统计舞段全部异读候选数。
func (s *VariantStore) CountByDance(danceID int64) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM variants WHERE dance_id=?`, danceID).Scan(&n)
	return n, err
}

// CountOpenByDance 统计未裁决候选数。
func (s *VariantStore) CountOpenByDance(danceID int64) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM variants WHERE dance_id=? AND status=?`,
		danceID, model.VariantStatusCandidate).Scan(&n)
	return n, err
}
