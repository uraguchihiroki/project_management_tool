package db

import (
	"fmt"

	"gorm.io/gorm"
)

// MigrateJunctionTablesSurrogatePK は複合主キーのみだった結合テーブルを id(UUID) 単独 PK に移行する（PostgreSQL 既存 DB 向け・冪等）。
// SQLite（テスト用メモリ DB 等）は新規 AutoMigrate で正しいスキーマが作られるため、ここでは何もしない。
func MigrateJunctionTablesSurrogatePK(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	for _, table := range []string{
		"user_roles",
		"organization_user_departments",
	} {
		if err := migratePostgresJunctionTable(db, table); err != nil {
			return err
		}
	}
	return nil
}

func migratePostgresJunctionTable(db *gorm.DB, table string) error {
	var n int64
	if err := db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = ? AND column_name = 'id'
	`, table).Scan(&n).Error; err != nil {
		return fmt.Errorf("junction %s: check id column: %w", table, err)
	}
	if n > 0 {
		return nil
	}
	var exists int64
	if err := db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = current_schema() AND table_name = ?
	`, table).Scan(&exists).Error; err != nil {
		return fmt.Errorf("junction %s: check table: %w", table, err)
	}
	if exists == 0 {
		return nil
	}

	var pkeyName string
	if err := db.Raw(`
		SELECT tc.constraint_name FROM information_schema.table_constraints tc
		WHERE tc.table_schema = current_schema() AND tc.table_name = ? AND tc.constraint_type = 'PRIMARY KEY'
	`, table).Scan(&pkeyName).Error; err != nil {
		return fmt.Errorf("junction %s: find pkey: %w", table, err)
	}
	if pkeyName == "" {
		return fmt.Errorf("junction %s: no primary key found", table)
	}

	q := func(stmt string) error {
		return db.Exec(stmt).Error
	}
	tq := quotePGIdent(table)
	pq := quotePGIdent(pkeyName)
	if err := q(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN id UUID`, tq)); err != nil {
		return fmt.Errorf("junction %s: add id: %w", table, err)
	}
	if err := q(fmt.Sprintf(`UPDATE %s SET id = gen_random_uuid() WHERE id IS NULL`, tq)); err != nil {
		return fmt.Errorf("junction %s: backfill id: %w", table, err)
	}
	if err := q(fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN id SET NOT NULL`, tq)); err != nil {
		return fmt.Errorf("junction %s: id not null: %w", table, err)
	}
	if err := q(fmt.Sprintf(`ALTER TABLE %s DROP CONSTRAINT %s`, tq, pq)); err != nil {
		return fmt.Errorf("junction %s: drop pkey: %w", table, err)
	}
	if err := q(fmt.Sprintf(`ALTER TABLE %s ADD PRIMARY KEY (id)`, tq)); err != nil {
		return fmt.Errorf("junction %s: add pkey id: %w", table, err)
	}
	// 複合検索用インデックスは直後の AutoMigrate（GORM タグ）に任せる
	return nil
}

func quotePGIdent(name string) string {
	return `"` + name + `"`
}
