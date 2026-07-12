package service

import (
	"fmt"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type CreditService struct {
	db *gorm.DB
}

func NewCreditService(db *gorm.DB) *CreditService {
	return &CreditService{db: db}
}

// CreateAccountWithTx 创建积分账户并赠送初始积分（事务中）
func (s *CreditService) CreateAccountWithTx(tx *gorm.DB, userID int64, initialAmount int) error {
	// 创建积分账户
	account := &model.CreditAccount{
		UserID:      userID,
		Balance:     initialAmount,
		TotalIncome: initialAmount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := tx.Create(account).Error; err != nil {
		return err
	}

	// 记录流水
	ledger := &model.CreditLedger{
		UserID:         userID,
		Type:           "income",
		Amount:         initialAmount,
		Title:          "新用户礼包",
		SourceType:     "gift",
		BalanceAfter:   initialAmount,
		IdempotencyKey: fmt.Sprintf("new_user_gift_%d_%d", userID, time.Now().Unix()),
		CreatedAt:      time.Now(),
	}
	return tx.Create(ledger).Error
}

// InvitationService 邀请服务
type InvitationService struct {
	db *gorm.DB
}

func NewInvitationService(db *gorm.DB) *InvitationService {
	return &InvitationService{db: db}
}

// ProcessRegisterRewardWithTx 处理注册邀请奖励（事务中）
func (s *InvitationService) ProcessRegisterRewardWithTx(tx *gorm.DB, inviterID, inviteeID int64, inviteCode string) error {
	now := time.Now()

	// 1. 创建邀请记录
	invitation := &model.Invitation{
		InviterID:    inviterID,
		InviteeID:    inviteeID,
		InviteCode:   inviteCode,
		Status:       "registered",
		RegisteredAt: now,
		CreatedAt:    now,
	}
	if err := tx.Create(invitation).Error; err != nil {
		return err
	}

	// 2. 读取奖励配置（这里硬编码，实际应从 system_configs 表读取）
	inviterReward := 50  // 邀请人奖励
	inviteeReward := 20  // 被邀请人奖励

	// 3. 发放奖励给邀请人
	if err := s.addCreditWithTx(tx, inviterID, inviterReward, "邀请奖励", "invite_register", invitation.ID); err != nil {
		return err
	}

	// 4. 发放奖励给被邀请人
	if err := s.addCreditWithTx(tx, inviteeID, inviteeReward, "注册奖励", "invite_register", invitation.ID); err != nil {
		return err
	}

	// 5. 记录奖励
	reward := &model.InvitationReward{
		InvitationID:  invitation.ID,
		InviterID:     inviterID,
		InviteeID:     inviteeID,
		InviterReward: inviterReward,
		InviteeReward: inviteeReward,
		TriggerType:   "register",
		CreatedAt:     now,
	}
	return tx.Create(reward).Error
}

// addCreditWithTx 增加积分（事务中）
func (s *InvitationService) addCreditWithTx(tx *gorm.DB, userID int64, amount int, title, sourceType string, sourceID int64) error {
	// 1. 更新账户余额
	var account model.CreditAccount
	if err := tx.Where("user_id = ?", userID).First(&account).Error; err != nil {
		return err
	}

	account.Balance += amount
	account.TotalIncome += amount
	account.UpdatedAt = time.Now()
	if err := tx.Save(&account).Error; err != nil {
		return err
	}

	// 2. 记录流水
	ledger := &model.CreditLedger{
		UserID:         userID,
		Type:           "income",
		Amount:         amount,
		Title:          title,
		SourceType:     sourceType,
		SourceID:       sourceID,
		BalanceAfter:   account.Balance,
		IdempotencyKey: fmt.Sprintf("%s_%d_%d_%d", sourceType, userID, sourceID, time.Now().Unix()),
		CreatedAt:      time.Now(),
	}
	return tx.Create(ledger).Error
}
