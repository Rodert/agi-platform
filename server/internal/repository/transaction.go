package repository

import (
	"context"

	"gorm.io/gorm"
)

type Tx interface {
	db() *gorm.DB
}

type gormTx struct {
	value *gorm.DB
}

func (tx gormTx) db() *gorm.DB {
	return tx.value
}

type TransactionManager interface {
	Transaction(ctx context.Context, fn func(tx Tx) error) error
}

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

func (m *GormTransactionManager) Transaction(ctx context.Context, fn func(tx Tx) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(gormTx{value: tx})
	})
}

func dbFromTx(tx Tx) *gorm.DB {
	if tx == nil {
		return nil
	}
	return tx.db()
}
