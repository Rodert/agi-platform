package repository

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

type DatabaseRepository interface {
	ListTables(ctx context.Context, schema string) ([]DatabaseTable, error)
	ListColumns(ctx context.Context, schema string, table string) ([]DatabaseColumn, error)
	ListRows(ctx context.Context, table string, limit int, offset int) (*DatabaseRows, error)
	GetDDL(ctx context.Context, table string) (string, error)
}

type DatabaseTable struct {
	Name      string `json:"name"`
	TableType string `json:"table_type"`
	Rows      int64  `gorm:"column:approx_rows" json:"rows"`
	Comment   string `json:"comment"`
}

type DatabaseColumn struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Nullable   bool   `json:"nullable"`
	PrimaryKey bool   `json:"primary_key"`
	Comment    string `json:"comment"`
}

type DatabaseRows struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
}

type GormDatabaseRepository struct {
	db *gorm.DB
}

var mysqlIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func NewGormDatabaseRepository(db *gorm.DB) *GormDatabaseRepository {
	return &GormDatabaseRepository{db: db}
}

func (r *GormDatabaseRepository) ListTables(ctx context.Context, schema string) ([]DatabaseTable, error) {
	var tables []DatabaseTable
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT TABLE_NAME AS name, TABLE_TYPE AS table_type, COALESCE(TABLE_ROWS, 0) AS approx_rows, TABLE_COMMENT AS comment
			FROM information_schema.TABLES
			WHERE TABLE_SCHEMA = ?
			ORDER BY TABLE_NAME ASC
		`, schema).
		Scan(&tables).
		Error
	return tables, err
}

func (r *GormDatabaseRepository) ListColumns(ctx context.Context, schema string, table string) ([]DatabaseColumn, error) {
	var raw []struct {
		Name      string `gorm:"column:name"`
		Type      string `gorm:"column:type"`
		Nullable  string `gorm:"column:nullable"`
		ColumnKey string `gorm:"column:column_key"`
		Comment   string `gorm:"column:comment"`
	}
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT COLUMN_NAME AS name, COLUMN_TYPE AS type, IS_NULLABLE AS nullable, COLUMN_KEY AS column_key, COLUMN_COMMENT AS comment
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION ASC
		`, schema, table).
		Scan(&raw).
		Error
	if err != nil {
		return nil, err
	}

	columns := make([]DatabaseColumn, 0, len(raw))
	for _, item := range raw {
		columns = append(columns, DatabaseColumn{
			Name:       item.Name,
			Type:       item.Type,
			Nullable:   strings.EqualFold(item.Nullable, "YES"),
			PrimaryKey: item.ColumnKey == "PRI",
			Comment:    item.Comment,
		})
	}
	return columns, nil
}

func (r *GormDatabaseRepository) GetDDL(ctx context.Context, table string) (string, error) {
	if !safeMySQLIdentifier(table) {
		return "", fmt.Errorf("invalid table name")
	}

	var row struct {
		Table       string `gorm:"column:Table"`
		CreateTable string `gorm:"column:Create Table"`
	}
	err := r.db.WithContext(ctx).Raw(fmt.Sprintf("SHOW CREATE TABLE `%s`", table)).Scan(&row).Error
	return row.CreateTable, err
}

func (r *GormDatabaseRepository) ListRows(ctx context.Context, table string, limit int, offset int) (*DatabaseRows, error) {
	if !safeMySQLIdentifier(table) {
		return nil, fmt.Errorf("invalid table name")
	}

	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT ? OFFSET ?", table)
	rows, err := sqlDB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		item := make(map[string]interface{}, len(columns))
		for i, column := range columns {
			item[column] = normalizeSQLValue(values[i])
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &DatabaseRows{Columns: columns, Rows: items}, nil
}

func safeMySQLIdentifier(value string) bool {
	return mysqlIdentifierPattern.MatchString(value)
}

func normalizeSQLValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case nil:
		return nil
	case []byte:
		return string(typed)
	case sql.RawBytes:
		return string(typed)
	default:
		return typed
	}
}
