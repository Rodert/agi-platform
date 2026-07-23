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
		responses = append(responses, &dto.AdminUserResponse{ID: user.ID, Email: user.Email, Name: user.Name, Level: user.Level, Balance: balance, CreatedAt: user.CreatedAt})
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
		if err := tx.Create(&model.CreditLedger{UserID: userID, Type: ledgerType, Amount: req.Amount, Title: title, SourceType: sourceType, SourceID: adminID, BalanceAfter: account.Balance, IdempotencyKey: "admin_adjustment_" + uuid.NewString(), CreatedAt: time.Now()}).Error; err != nil { return err }
		beforeData, _ := json.Marshal(map[string]int{"balance": before})
		afterData, _ := json.Marshal(map[string]interface{}{"balance": account.Balance, "type": req.Type, "amount": req.Amount, "remark": req.Remark})
		if err := tx.Create(&model.AdminLog{AdminID: adminID, Action: "adjust_credit", TargetType: "user", TargetID: userID, BeforeData: string(beforeData), AfterData: string(afterData), Description: title + "：" + req.Remark, CreatedAt: time.Now()}).Error; err != nil { return err }
		result.Balance = account.Balance
		return nil
	})
	if err != nil { return nil, err }
	return result, nil
}

// CreateUser 创建用户。
func (s *AdminService) CreateUser(req *dto.CreateUserRequest) error {
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
	return s.userRepo.Create(&model.User{
		Email: req.Email, PasswordHash: hashedPassword, Name: req.Username,
		Level: "free", InviteCode: inviteCode, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *AdminService) UpdateUser(userID int64, req *dto.AdminUpdateUserRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound { return errors.ErrUserNotFound }
		return err
	}
	if req.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmailExceptID(req.Email, userID)
		if err != nil { return err }
		if exists { return errors.ErrUserExists }
	}
	user.Name = req.Username
	user.Email = req.Email
	user.Level = req.Level
	if req.Password != "" {
		hash, err := utils.HashPassword(req.Password)
		if err != nil { return err }
		user.PasswordHash = hash
	}
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(user)
}

// UpdateUserStatus 当前用户表没有启停字段。
func (s *AdminService) UpdateUserStatus(userID int64, _ bool) error {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		return err
	}
	return errors.New(errors.ErrCodeBadRequest, "当前用户模型暂不支持启停状态")
}
