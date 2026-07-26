package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type CreditRepository struct {
	db *gorm.DB
}

func (r *CreditRepository) CreateRedeemCodes(tx *gorm.DB, codes []*model.RedeemCode) error {
	return tx.Create(&codes).Error
}

func (r *CreditRepository) ListActivePackages() ([]*model.CreditPackage, error) {
	var items []*model.CreditPackage
	err := r.db.Where("is_active = ?", true).Order("sort_order ASC, id ASC").Limit(3).Find(&items).Error
	return items, err
}

func (r *CreditRepository) ListPackages() ([]*model.CreditPackage, error) {
	var items []*model.CreditPackage
	err := r.db.Order("sort_order ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *CreditRepository) UpdatePackages(packages []*model.CreditPackage) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range packages {
			updates := map[string]interface{}{
				"name": item.Name, "price": item.Price, "points": item.Points,
				"note": item.Note, "purchase_url": item.PurchaseURL, "is_hot": item.IsHot,
				"is_active": item.IsActive, "sort_order": item.SortOrder,
			}
			result := tx.Model(&model.CreditPackage{}).Where("id = ?", item.ID).Updates(updates)
			if result.Error != nil { return result.Error }
		}
		return nil
	})
}

func (r *CreditRepository) ListRedeemCodes(keyword, status string, page, pageSize int) ([]*model.RedeemCode, int64, error) {
	query := r.db.Model(&model.RedeemCode{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		query = query.Where("code LIKE ? OR batch_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	now := time.Now()
	switch status {
	case "used": query = query.Where("used_by <> 0")
	case "unused": query = query.Where("used_by = 0 AND (expires_at IS NULL OR expires_at >= ?)", now)
	case "expired": query = query.Where("used_by = 0 AND expires_at IS NOT NULL AND expires_at < ?", now)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	var codes []*model.RedeemCode
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&codes).Error
	return codes, total, err
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

// RefundFailedTask restores exactly the expense recorded for a task. The
// unique refund key and account lock make failure handling idempotent.
func (r *CreditRepository) RefundFailedTask(task *model.Task) error {
	if task.ID == 0 || task.UserID == 0 {
		return fmt.Errorf("无效任务，不能返还灵感值")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var expense model.CreditLedger
		expenseKey := fmt.Sprintf("task_expense_%d", task.ID)
		if err := tx.Where("user_id = ? AND type = ? AND source_type = ? AND source_id = ? AND idempotency_key = ?", task.UserID, "expense", "task", task.ID, expenseKey).First(&expense).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("未找到任务 %d 对应的扣费流水", task.ID)
			}
			return err
		}
		if expense.Amount <= 0 || expense.Amount != task.Cost {
			return fmt.Errorf("任务 %d 的扣费流水校验失败", task.ID)
		}

		account, err := r.GetAccountForUpdate(tx, task.UserID)
		if err != nil {
			return err
		}
		refundKey := fmt.Sprintf("task_refund_%d", task.ID)
		var refunded model.CreditLedger
		err = tx.Where("idempotency_key = ?", refundKey).First(&refunded).Error
		if err == nil {
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		account.Balance += expense.Amount
		account.TotalIncome += expense.Amount
		account.UpdatedAt = time.Now()
		if err := r.UpdateAccount(tx, account); err != nil {
			return err
		}
		return r.CreateLedger(tx, &model.CreditLedger{
			UserID:         task.UserID,
			Type:           "income",
			Amount:         expense.Amount,
			Title:          "生成任务失败返还",
			SourceType:     "task_refund",
			SourceID:       expense.ID,
			BalanceAfter:   account.Balance,
			IdempotencyKey: refundKey,
			CreatedAt:      time.Now(),
		})
	})
}

// GetLedgers 获取流水列表
func (r *CreditRepository) GetLedgers(userID int64, ledgerType, sourceType string, startAt, endAt *time.Time, page, pageSize int) ([]*model.CreditLedger, int64, error) {
	var ledgers []*model.CreditLedger
	var total int64

	query := r.db.Model(&model.CreditLedger{}).Where("user_id = ?", userID)
	if ledgerType != "" { query = query.Where("type = ?", ledgerType) }
	if sourceType != "" { query = query.Where("source_type = ?", sourceType) }
	if startAt != nil { query = query.Where("created_at >= ?", *startAt) }
	if endAt != nil { query = query.Where("created_at <= ?", *endAt) }

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
