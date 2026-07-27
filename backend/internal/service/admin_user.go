package service

import (
	"encoding/json"
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetUserList 获取用户列表。
func (s *AdminService) GetUserList(req *dto.UserListRequest) (*dto.UserListResponse, error) {
	users, total, err := s.userRepo.GetList(req.Page, req.PageSize, req.Username, req.Email)
	if err != nil {
		return nil, errors.NewWithDetails(errors.ErrCodeInternalServer, "查询用户列表失败", err.Error())
	}
	responses := make([]*dto.AdminUserResponse, 0, len(users))
	for _, user := range users {
		balance := 0
		account, accountErr := s.creditRepo.GetAccount(user.ID)
		if accountErr != nil && accountErr != gorm.ErrRecordNotFound { return nil, accountErr }
		if accountErr == nil { balance = account.Balance }
		responses = append(responses, &dto.AdminUserResponse{ID: user.ID, Email: user.Email, Name: user.Name, Level: user.Level, IsActive: user.IsActive, Balance: balance, CreatedAt: user.CreatedAt})
	}
	return &dto.UserListResponse{List: responses, Total: total}, nil
}

func (s *AdminService) RechargeUserCredit(adminID, userID int64, req *dto.AdminRechargeCreditRequest) (*dto.AdminRechargeCreditResponse, error) {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		if err == gorm.ErrRecordNotFound { return nil, errors.ErrUserNotFound }
		return nil, err
	}
	result := &dto.AdminRechargeCreditResponse{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var account model.CreditAccount
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&account).Error
		if err == gorm.ErrRecordNotFound {
			account = model.CreditAccount{UserID: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			if err := tx.Create(&account).Error; err != nil { return err }
		} else if err != nil { return err }
		before := account.Balance
		ledgerType, sourceType, title := "income", "admin_adjustment_add", "管理员增加灵感值"
		if req.Type == "deduct" {
			if account.Balance < req.Amount { return errors.ErrInsufficientCredit }
			account.Balance -= req.Amount
			account.TotalExpense += req.Amount
			ledgerType, sourceType, title = "expense", "admin_adjustment_deduct", "管理员扣减灵感值"
		} else {
			account.Balance += req.Amount
			account.TotalIncome += req.Amount
		}
		account.UpdatedAt = time.Now()
		if err := tx.Save(&account).Error; err != nil { return err }
		ledgerTitle := title + "：" + req.Remark
		if err := tx.Create(&model.CreditLedger{UserID: userID, Type: ledgerType, Amount: req.Amount, Title: ledgerTitle, SourceType: sourceType, SourceID: adminID, BalanceAfter: account.Balance, IdempotencyKey: "admin_adjustment_" + uuid.NewString(), CreatedAt: time.Now()}).Error; err != nil { return err }
		beforeData, _ := json.Marshal(map[string]int{"balance": before})
		afterData, _ := json.Marshal(map[string]interface{}{"balance": account.Balance, "type": req.Type, "amount": req.Amount, "remark": req.Remark})
		if err := tx.Create(&model.AdminLog{AdminID: adminID, Action: "adjust_credit", TargetType: "user", TargetID: userID, BeforeData: string(beforeData), AfterData: string(afterData), Description: ledgerTitle, CreatedAt: time.Now()}).Error; err != nil { return err }
		result.Balance = account.Balance
		return nil
	})
	if err != nil { return nil, err }
	return result, nil
}

// GetUserCreditLedgers returns the immutable credit history for one user.
func (s *AdminService) GetUserCreditLedgers(userID int64, req *dto.AdminCreditLedgerListRequest) ([]*dto.AdminCreditLedgerResponse, int64, error) {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, errors.ErrUserNotFound
		}
		return nil, 0, err
	}
	startAt, err := parseAdminLogTime(req.StartAt)
	if err != nil { return nil, 0, err }
	endAt, err := parseAdminLogTime(req.EndAt)
	if err != nil { return nil, 0, err }
	if startAt != nil && endAt != nil && endAt.Before(*startAt) {
		return nil, 0, errors.New(errors.ErrCodeBadRequest, "结束时间不能早于开始时间")
	}
	ledgers, total, err := s.creditRepo.GetLedgers(userID, req.Type, req.SourceType, startAt, endAt, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*dto.AdminCreditLedgerResponse, 0, len(ledgers))
	for _, ledger := range ledgers {
		items = append(items, &dto.AdminCreditLedgerResponse{
			ID: ledger.ID, Type: ledger.Type, Amount: ledger.Amount, Title: ledger.Title,
			SourceType: ledger.SourceType, SourceID: ledger.SourceID,
			BalanceAfter: ledger.BalanceAfter, CreatedAt: ledger.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return items, total, nil
}

// CreateUser 创建用户。
func (s *AdminService) CreateUser(adminID int64, req *dto.CreateUserRequest) error {
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.ErrUserExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	var inviteCode string
	for i := 0; i < 10; i++ {
		candidate := utils.GenerateInviteCode()
		exists, err := s.userRepo.ExistsByInviteCode(candidate)
		if err != nil {
			return err
		}
		if !exists {
			inviteCode = candidate
			break
		}
	}
	if inviteCode == "" {
		return errors.New(errors.ErrCodeInternalServer, "生成邀请码失败")
	}
	now := time.Now()
	user := &model.User{
		Email: req.Email, PasswordHash: hashedPassword, Name: req.Username,
		Level: "free", IsActive: true, InviteCode: inviteCode, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.userRepo.Create(user); err != nil {
		return err
	}
	afterData, _ := json.Marshal(map[string]interface{}{"email": user.Email, "name": user.Name, "level": user.Level})
	s.recordAudit(&model.AdminLog{AdminID: adminID, Action: "create_user", TargetType: "user", TargetID: user.ID, AfterData: string(afterData), Description: "创建用户", CreatedAt: time.Now()})
	return nil
}

func (s *AdminService) UpdateUser(adminID, userID int64, req *dto.AdminUpdateUserRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound { return errors.ErrUserNotFound }
		return err
	}
	beforeData, _ := json.Marshal(map[string]interface{}{"email": user.Email, "name": user.Name, "level": user.Level})
	user.Name = req.Username
	user.Level = req.Level
	if req.Password != "" {
		hash, err := utils.HashPassword(req.Password)
		if err != nil { return err }
		user.PasswordHash = hash
	}
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	afterData, _ := json.Marshal(map[string]interface{}{"email": user.Email, "name": user.Name, "level": user.Level, "password_changed": req.Password != ""})
	s.recordAudit(&model.AdminLog{AdminID: adminID, Action: "update_user", TargetType: "user", TargetID: userID, BeforeData: string(beforeData), AfterData: string(afterData), Description: "更新用户", CreatedAt: time.Now()})
	return nil
}

// UpdateUserStatus changes a user's access state. Disabling also invalidates
// all existing sessions so the account loses access immediately.
func (s *AdminService) UpdateUserStatus(adminID, userID int64, isActive bool) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		return err
	}
	beforeData, _ := json.Marshal(map[string]bool{"is_active": user.IsActive})
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"is_active": isActive,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}
		if !isActive {
			if err := tx.Model(&model.UserSession{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", time.Now()).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	afterData, _ := json.Marshal(map[string]bool{"is_active": isActive})
	action, description := "enable_user", "启用用户"
	if !isActive {
		action, description = "disable_user", "停用用户并撤销全部登录会话"
	}
	s.recordAudit(&model.AdminLog{AdminID: adminID, Action: action, TargetType: "user", TargetID: userID, BeforeData: string(beforeData), AfterData: string(afterData), Description: description, CreatedAt: time.Now()})
	return nil
}
