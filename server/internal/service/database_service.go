package service

import (
	"context"
	"strings"

	"agi-platform/server/internal/repository"
)

type DatabaseExplorerService interface {
	ListTables(ctx context.Context) ([]repository.DatabaseTable, error)
	GetTable(ctx context.Context, table string, limit int, offset int) (*DatabaseTableData, error)
}

type DatabaseTableData struct {
	Table   string                      `json:"table"`
	Comment string                      `json:"comment"`
	Columns []repository.DatabaseColumn `json:"columns"`
	Rows    []map[string]interface{}    `json:"rows"`
	DDL     string                      `json:"ddl"`
	Limit   int                         `json:"limit"`
	Offset  int                         `json:"offset"`
	HasNext bool                        `json:"has_next"`
}

type databaseExplorerService struct {
	repos  repository.Repositories
	schema string
}

func NewDatabaseExplorerService(repos repository.Repositories, schema string) DatabaseExplorerService {
	return &databaseExplorerService{
		repos:  repos,
		schema: strings.TrimSpace(schema),
	}
}

func (s *databaseExplorerService) ListTables(ctx context.Context) ([]repository.DatabaseTable, error) {
	if s.schema == "" {
		return nil, ErrInvalidRequest
	}
	tables, err := s.repos.Database.ListTables(ctx, s.schema)
	if err != nil {
		return nil, err
	}
	enrichDatabaseTableComments(tables)
	return tables, nil
}

func (s *databaseExplorerService) GetTable(ctx context.Context, table string, limit int, offset int) (*DatabaseTableData, error) {
	table = strings.TrimSpace(table)
	if table == "" || s.schema == "" {
		return nil, ErrInvalidRequest
	}

	columns, err := s.repos.Database.ListColumns(ctx, s.schema, table)
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, repository.ErrNotFound
	}
	enrichDatabaseColumnComments(columns)

	limit, offset = normalizePage(limit, offset)
	rows, err := s.repos.Database.ListRows(ctx, table, limit+1, offset)
	if err != nil {
		return nil, err
	}
	hasNext := len(rows.Rows) > limit
	if hasNext {
		rows.Rows = rows.Rows[:limit]
	}
	ddl, err := s.repos.Database.GetDDL(ctx, table)
	if err != nil {
		return nil, err
	}

	return &DatabaseTableData{
		Table:   table,
		Comment: databaseTableComments[table],
		Columns: columns,
		Rows:    rows.Rows,
		DDL:     ddl,
		Limit:   limit,
		Offset:  offset,
		HasNext: hasNext,
	}, nil
}
