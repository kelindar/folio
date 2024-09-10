package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/kelindar/folio"
)

func autoMigrate(db *sql.DB, registry folio.Registry) error {
	if err := registry.Range(func(k folio.Kind, _ reflect.Type) error {
		return errors.Join(
			createTable(db, tableOf(k)),
			createSearchIndex(db, tableOf(k)),
		)
	}); err != nil {
		return err
	}

	return nil
}

func createTable(db *sql.DB, table string) error {
	return errors.Join(
		execf(db, `CREATE TABLE IF NOT EXISTS %s ( id TEXT PRIMARY KEY, namespace TEXT, state TEXT, data JSON, indexed_by TEXT, created_by TEXT, updated_by TEXT, created_at INTEGER, updated_at INTEGER)`, table),
		execf(db, `CREATE INDEX IF NOT EXISTS %s_idx_namespace ON %s(namespace)`, table, table),
		execf(db, `CREATE INDEX IF NOT EXISTS %s_idx_state ON %s(state)`, table, table),
		execf(db, `CREATE INDEX IF NOT EXISTS %s_idx_index ON %s(indexed_by)`, table, table),
	)
}

func createSearchIndex(db *sql.DB, table string) error {
	return errors.Join(
		execf(db, `CREATE VIRTUAL TABLE IF NOT EXISTS %s_fts USING fts5(id, data)`,
			table),
		execf(db, `CREATE TRIGGER IF NOT EXISTS %s_fts_before_update BEFORE UPDATE ON %s BEGIN DELETE FROM %s_fts WHERE rowid = old.rowid; END`,
			table, table, table),
		execf(db, `CREATE TRIGGER IF NOT EXISTS %s_fts_before_delete BEFORE DELETE ON %s BEGIN DELETE FROM %s_fts WHERE rowid = old.rowid; END`,
			table, table, table),
		execf(db, `CREATE TRIGGER IF NOT EXISTS %s_after_update AFTER UPDATE ON %s BEGIN INSERT INTO %s_fts(rowid, id, data) VALUES (new.rowid, new.id, new.data); END`,
			table, table, table),
		execf(db, `CREATE TRIGGER IF NOT EXISTS %s_after_insert AFTER INSERT ON %s BEGIN INSERT INTO %s_fts(rowid, id, data) VALUES (new.rowid, new.id, new.data); END`,
			table, table, table),
	)
}

func execf(db *sql.DB, sql string, args ...any) error {
	_, err := db.Exec(fmt.Sprintf(sql, args...))
	return err
}
