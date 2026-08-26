// SQLite 持久化层：建表迁移与仓储访问。
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// Store 聚合全部仓储，共享同一个 SQLite 连接。
type Store struct {
	db *sql.DB
	mu sync.Mutex

	Dances     *DanceStore
	Movements  *MovementStore
	Beats      *BeatStore
	Formations *FormationStore
	Variants   *VariantStore
	Versions   *VersionStore
}

// Open 打开（或创建）SQLite 数据库并执行迁移。
func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("empty db path")
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir db dir: %w", err)
		}
	}
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite 单写者，串行化连接避免锁竞争
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	s := &Store{db: db}
	s.Dances = &DanceStore{db: db, mu: &s.mu}
	s.Movements = &MovementStore{db: db}
	s.Beats = &BeatStore{db: db}
	s.Formations = &FormationStore{db: db}
	s.Variants = &VariantStore{db: db}
	s.Versions = &VersionStore{db: db}
	return s, nil
}

// Close 关闭数据库。
func (s *Store) Close() error {
	return s.db.Close()
}

// WithTx 在单个事务内执行 fn，成功提交、失败回滚。
func (s *Store) WithTx(fn func(tx *sql.Tx) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// migrate 建表迁移（幂等）。
func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS dances (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			region TEXT NOT NULL DEFAULT '',
			dancers INTEGER NOT NULL,
			status TEXT NOT NULL,
			notes TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_dances_name ON dances(name)`,

		`CREATE TABLE IF NOT EXISTS movements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dance_id INTEGER NOT NULL REFERENCES dances(id),
			dancer_no INTEGER NOT NULL,
			action_name TEXT NOT NULL,
			start_beat INTEGER NOT NULL,
			end_beat INTEGER NOT NULL,
			connection TEXT NOT NULL DEFAULT 'continuous',
			status TEXT NOT NULL,
			fingerprint TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE(dance_id, dancer_no, start_beat)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_movements_fp ON movements(fingerprint)`,

		`CREATE TABLE IF NOT EXISTS beat_anchors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dance_id INTEGER NOT NULL REFERENCES dances(id),
			beat_no INTEGER NOT NULL,
			image_time REAL NOT NULL,
			confidence REAL NOT NULL,
			source TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			UNIQUE(dance_id, beat_no)
		)`,

		`CREATE TABLE IF NOT EXISTS formations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dance_id INTEGER NOT NULL REFERENCES dances(id),
			from_dancer INTEGER NOT NULL,
			to_dancer INTEGER NOT NULL,
			relation TEXT NOT NULL,
			beat_start INTEGER NOT NULL,
			beat_end INTEGER NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE(dance_id, from_dancer, to_dancer, beat_start)
		)`,

		`CREATE TABLE IF NOT EXISTS variants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dance_id INTEGER NOT NULL REFERENCES dances(id),
			type TEXT NOT NULL,
			ref_id INTEGER NOT NULL,
			detail TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			verdict TEXT NOT NULL DEFAULT '',
			reason TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			decided_at DATETIME
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_variants_ref ON variants(dance_id, type, ref_id)`,

		`CREATE TABLE IF NOT EXISTS notation_versions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dance_id INTEGER NOT NULL REFERENCES dances(id),
			version_no INTEGER NOT NULL,
			status TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			evidence TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			UNIQUE(dance_id, version_no)
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
