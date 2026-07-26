package repository

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByID 根据ID查找用户
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByInviteCode 根据邀请码查找用户
func (r *UserRepository) FindByInviteCode(code string) (*model.User, error) {
	var user model.User
	err := r.db.Where("invite_code = ?", code).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByPhoneExceptID(phone string, id int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("phone = ? AND id <> ?", phone, id).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByEmailExceptID(email string, id int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ? AND id <> ?", email, id).Count(&count).Error
	return count > 0, err
}

// ExistsByInviteCode 检查邀请码是否存在
func (r *UserRepository) ExistsByInviteCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("invite_code = ?", code).Count(&count).Error
	return count > 0, err
}

// GetList 返回管理后台用户列表。
func (r *UserRepository) GetList(page, pageSize int, username, email string) ([]*model.User, int64, error) {
	query := r.db.Model(&model.User{})
	if username != "" {
		query = query.Where("name LIKE ?", "%"+username+"%")
	}
	if email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []*model.User
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// VerificationCodeRepository 验证码仓库
type VerificationCodeRepository struct {
	db *gorm.DB
}

func NewVerificationCodeRepository(db *gorm.DB) *VerificationCodeRepository {
	return &VerificationCodeRepository{db: db}
}

// Create 创建验证码
func (r *VerificationCodeRepository) Create(code *model.VerificationCode) error {
	return r.db.Create(code).Error
}

// FindLatest 查找最新的验证码
func (r *VerificationCodeRepository) FindLatest(email, codeType string) (*model.VerificationCode, error) {
	var code model.VerificationCode
	err := r.db.Where("email = ? AND type = ? AND used_at IS NULL", email, codeType).
		Order("created_at DESC").
		First(&code).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// FindActive returns the most recently issued usable code for an address and purpose.
func (r *VerificationCodeRepository) FindActive(email, codeType string) (*model.VerificationCode, error) {
	var code model.VerificationCode
	err := r.db.Where("email = ? AND type = ? AND used_at IS NULL AND expires_at > ?", email, codeType, time.Now()).
		Order("created_at DESC").
		First(&code).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// FindMatchingActive finds a usable code by its value instead of assuming the
// latest email arrived first.
func (r *VerificationCodeRepository) FindMatchingActive(email, value, codeType string) (*model.VerificationCode, error) {
	var code model.VerificationCode
	err := r.db.Where("email = ? AND code = ? AND type = ? AND used_at IS NULL", email, value, codeType).
		Order("created_at DESC").
		First(&code).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// MarkAsUsed 标记验证码为已使用
func (r *VerificationCodeRepository) MarkAsUsed(id int64) error {
	now := time.Now()
	return r.db.Model(&model.VerificationCode{}).
		Where("id = ?", id).
		Update("used_at", now).Error
}

// DeleteExpired 删除过期的验证码
func (r *VerificationCodeRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ? AND used_at IS NULL", time.Now()).
		Delete(&model.VerificationCode{}).Error
}

type UserSessionRepository struct { db *gorm.DB }

func NewUserSessionRepository(db *gorm.DB) *UserSessionRepository { return &UserSessionRepository{db: db} }

func (r *UserSessionRepository) Create(session *model.UserSession) error { return r.db.Create(session).Error }

func (r *UserSessionRepository) FindActive(id string) (*model.UserSession, error) {
	var session model.UserSession
	err := r.db.Where("id = ? AND revoked_at IS NULL AND expires_at > ?", id, time.Now()).First(&session).Error
	return &session, err
}

func (r *UserSessionRepository) ListActive(userID int64) ([]*model.UserSession, error) {
	var sessions []*model.UserSession
	err := r.db.Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).Order("updated_at DESC").Find(&sessions).Error
	return sessions, err
}

func (r *UserSessionRepository) Revoke(userID int64, id string) error {
	now := time.Now()
	return r.db.Model(&model.UserSession{}).Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).Update("revoked_at", now).Error
}

func (r *UserSessionRepository) RevokeOther(userID int64, currentID string) error {
	now := time.Now()
	return r.db.Model(&model.UserSession{}).Where("user_id = ? AND id <> ? AND revoked_at IS NULL", userID, currentID).Update("revoked_at", now).Error
}
