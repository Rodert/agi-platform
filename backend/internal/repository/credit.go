package repository

import (
	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type CreditRepository struct {
	db *gorm.DB
}

func NewCreditRepository(db *gorm.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

// GetAccount 获取积分账户
func (r *CreditRepository) GetAccount(userID int64) (*model.CreditAccount, error) {
	var account model.CreditAccount
	err := r.db.Where("user_id = ?", userID).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetOrCreateAccount guarantees every user has a persisted credit account.
func (r *CreditRepository) GetOrCreateAccount(userID int64) (*model.CreditAccount, error) {
	account := &model.CreditAccount{UserID: userID}
	if err := r.db.Where("user_id = ?", userID).FirstOrCreate(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

// GetAccountForUpdate 获取积分账户（加锁）
func (r *CreditRepository) GetAccountForUpdate(tx *gorm.DB, userID int64) (*model.CreditAccount, error) {
	var account model.CreditAccount
	err := tx.Raw("SELECT * FROM credit_accounts WHERE user_id = ? FOR UPDATE", userID).
		Scan(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// UpdateAccount 更新账户
func (r *CreditRepository) UpdateAccount(tx *gorm.DB, account *model.CreditAccount) error {
	return tx.Save(account).Error
}

// CreateLedger 创建流水记录
func (r *CreditRepository) CreateLedger(tx *gorm.DB, ledger *model.CreditLedger) error {
	return tx.Create(ledger).Error
}

// GetLedgers 获取流水列表
func (r *CreditRepository) GetLedgers(userID int64, page, pageSize int) ([]*model.CreditLedger, int64, error) {
	var ledgers []*model.CreditLedger
	var total int64

	query := r.db.Model(&model.CreditLedger{}).Where("user_id = ?", userID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&ledgers).Error

	return ledgers, total, err
}
