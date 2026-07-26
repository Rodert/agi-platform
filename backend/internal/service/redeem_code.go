package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RedeemCodeService struct {
	db         *gorm.DB
	creditRepo *repository.CreditRepository
}

func NewRedeemCodeService(db *gorm.DB, creditRepo *repository.CreditRepository) *RedeemCodeService {
	return &RedeemCodeService{db: db, creditRepo: creditRepo}
}

func (s *RedeemCodeService) ListActivePackages() ([]*dto.CreditPackageResponse, error) {
	packages, err := s.creditRepo.ListActivePackages()
	if err != nil { return nil, err }
	result := make([]*dto.CreditPackageResponse, 0, len(packages))
	for _, item := range packages { result = append(result, &dto.CreditPackageResponse{ID: item.ID, Name: item.Name, Price: item.Price, Points: item.Points, Note: item.Note, PurchaseURL: item.PurchaseURL, IsHot: item.IsHot}) }
	return result, nil
}

func (s *RedeemCodeService) ListPackages() ([]*model.CreditPackage, error) { return s.creditRepo.ListPackages() }

func (s *RedeemCodeService) UpdatePackages(req []*dto.AdminCreditPackageRequest) error {
	if len(req) != 3 { return errors.New(errors.ErrCodeBadRequest, "必须配置三个套餐") }
	packages := make([]*model.CreditPackage, 0, len(req))
	seen := map[int64]bool{}
	for index, item := range req {
		if seen[item.ID] { return errors.New(errors.ErrCodeBadRequest, "套餐不能重复") }
		seen[item.ID] = true
		packages = append(packages, &model.CreditPackage{ID: item.ID, Name: strings.TrimSpace(item.Name), Price: item.Price, Points: item.Points, Note: strings.TrimSpace(item.Note), PurchaseURL: strings.TrimSpace(item.PurchaseURL), IsHot: item.IsHot, IsActive: item.IsActive, SortOrder: index + 1})
	}
	return s.creditRepo.UpdatePackages(packages)
}

func (s *RedeemCodeService) Create(adminID int64, req *dto.AdminCreateRedeemCodesRequest) ([]*dto.AdminRedeemCodeResponse, error) {
	batchID := "RC-" + strings.ToUpper(uuid.NewString()[:8])
	batchName := strings.TrimSpace(req.BatchName)
	if batchName == "" { batchName = defaultRedeemBatchName() }
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		value, err := time.ParseInLocation("2006-01-02 15:04:05", req.ExpiresAt, time.Local)
		if err != nil { return nil, errors.New(errors.ErrCodeBadRequest, "有效期格式错误") }
		if !value.After(time.Now()) { return nil, errors.New(errors.ErrCodeBadRequest, "有效期必须晚于当前时间") }
		expiresAt = &value
	}
	now := time.Now()
	codes := make([]*model.RedeemCode, 0, req.Quantity)
	for len(codes) < req.Quantity {
		code, err := generateRedeemCode()
		if err != nil { return nil, err }
		codes = append(codes, &model.RedeemCode{Code: code, Amount: req.Amount, BatchID: batchID, BatchName: batchName, ExpiresAt: expiresAt, CreatedAt: now})
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.creditRepo.CreateRedeemCodes(tx, codes); err != nil { return err }
		return tx.Create(&model.AdminLog{AdminID: adminID, Action: "create_redeem_codes", TargetType: "redeem_code_batch", AfterData: fmt.Sprintf(`{"batch_id":%q,"batch_name":%q,"amount":%d,"quantity":%d}`, batchID, batchName, req.Amount, req.Quantity), Description: "创建兑换码批次：" + batchName, CreatedAt: now}).Error
	}); err != nil { return nil, err }
	result := make([]*dto.AdminRedeemCodeResponse, 0, len(codes))
	for _, code := range codes { result = append(result, redeemCodeResponse(code)) }
	return result, nil
}

func defaultRedeemBatchName() string {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil { return "AGI-" + strings.ToUpper(uuid.NewString()[:6]) }
	return "AGI-" + strings.ToUpper(hex.EncodeToString(bytes))
}

func (s *RedeemCodeService) Redeem(userID int64, rawCode string) (*dto.RedeemCodeResponse, error) {
	codeValue := strings.ToUpper(strings.TrimSpace(rawCode))
	if codeValue == "" { return nil, errors.ErrInvalidRedeemCode }
	result := &dto.RedeemCodeResponse{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var code model.RedeemCode
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code = ?", codeValue).First(&code).Error; err != nil {
			if err == gorm.ErrRecordNotFound { return errors.ErrInvalidRedeemCode }
			return err
		}
		if code.UsedBy != 0 { return errors.ErrRedeemCodeUsed }
		now := time.Now()
		if code.ExpiresAt != nil && !code.ExpiresAt.After(now) { return errors.ErrRedeemCodeExpired }
		var account model.CreditAccount
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&account).Error
		if err == gorm.ErrRecordNotFound {
			account = model.CreditAccount{UserID: userID, CreatedAt: now, UpdatedAt: now}
			if err := tx.Create(&account).Error; err != nil { return err }
		} else if err != nil { return err }
		account.Balance += code.Amount
		account.TotalIncome += code.Amount
		account.UpdatedAt = now
		if err := tx.Save(&account).Error; err != nil { return err }
		code.UsedBy, code.UsedAt = userID, &now
		if err := tx.Save(&code).Error; err != nil { return err }
		if err := tx.Create(&model.CreditLedger{UserID: userID, Type: "income", Amount: code.Amount, Title: "兑换码：" + code.BatchName, SourceType: "redeem_code", SourceID: code.ID, BalanceAfter: account.Balance, IdempotencyKey: "redeem_code_" + fmt.Sprint(code.ID), CreatedAt: now}).Error; err != nil { return err }
		result.Amount, result.Balance = code.Amount, account.Balance
		return nil
	})
	if err != nil { return nil, err }
	return result, nil
}

func (s *RedeemCodeService) List(req *dto.AdminRedeemCodeListRequest) ([]*dto.AdminRedeemCodeResponse, int64, error) {
	codes, total, err := s.creditRepo.ListRedeemCodes(req.Keyword, req.Status, req.Page, req.PageSize)
	if err != nil { return nil, 0, err }
	userIDs := make([]int64, 0)
	for _, code := range codes { if code.UsedBy != 0 { userIDs = append(userIDs, code.UsedBy) } }
	users := map[int64]model.User{}
	if len(userIDs) > 0 {
		var rows []model.User
		if err := s.db.Where("id IN ?", userIDs).Find(&rows).Error; err != nil { return nil, 0, err }
		for _, user := range rows { users[user.ID] = user }
	}
	items := make([]*dto.AdminRedeemCodeResponse, 0, len(codes))
	for _, code := range codes {
		item := redeemCodeResponse(code)
		if user, ok := users[code.UsedBy]; ok { item.UsedName, item.UsedEmail = user.Name, user.Email }
		items = append(items, item)
	}
	return items, total, nil
}

func generateRedeemCode() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil { return "", err }
	encoded := strings.ToUpper(hex.EncodeToString(bytes))
	return "AGI-" + encoded[:4] + "-" + encoded[4:8] + "-" + encoded[8:], nil
}

func redeemCodeResponse(code *model.RedeemCode) *dto.AdminRedeemCodeResponse {
	item := &dto.AdminRedeemCodeResponse{ID: code.ID, Code: code.Code, Amount: code.Amount, BatchID: code.BatchID, BatchName: code.BatchName, UsedBy: code.UsedBy, CreatedAt: code.CreatedAt.Format("2006-01-02 15:04:05")}
	if code.UsedAt != nil { item.UsedAt = code.UsedAt.Format("2006-01-02 15:04:05") }
	if code.ExpiresAt != nil { item.ExpiresAt = code.ExpiresAt.Format("2006-01-02 15:04:05") }
	return item
}
