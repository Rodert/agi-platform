package service

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"gorm.io/gorm"
)

func adminManagerResponse(admin *model.AdminUser) *dto.AdminManagerResponse {
	item := &dto.AdminManagerResponse{ID: admin.ID, Username: admin.Username, Name: admin.Name, Role: admin.Role, IsActive: admin.IsActive, CreatedAt: admin.CreatedAt.Format(adminLogDateLayout)}
	if admin.LastLoginAt != nil { item.LastLoginAt = admin.LastLoginAt.Format(adminLogDateLayout) }
	return item
}

func (s *AdminService) ListAdmins() ([]*dto.AdminManagerResponse, error) {
	admins, err := s.adminRepo.ListAdmins()
	if err != nil { return nil, err }
	items := make([]*dto.AdminManagerResponse, 0, len(admins))
	for _, admin := range admins { items = append(items, adminManagerResponse(admin)) }
	return items, nil
}

func (s *AdminService) CreateAdmin(operatorID int64, req *dto.CreateAdminManagerRequest) (*dto.AdminManagerResponse, error) {
	if _, err := s.adminRepo.FindAdminByUsername(req.Username); err == nil { return nil, errors.New(errors.ErrCodeBadRequest, "用户名已存在") } else if err != gorm.ErrRecordNotFound { return nil, err }
	hash, err := utils.HashPassword(req.Password)
	if err != nil { return nil, err }
	admin := &model.AdminUser{Username: req.Username, Name: req.Name, PasswordHash: hash, Role: req.Role, IsActive: true, Permissions: "[]", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.adminRepo.CreateAdmin(admin); err != nil { return nil, err }
	s.recordAudit(&model.AdminLog{AdminID: operatorID, Action: "create_admin", TargetType: "admin", TargetID: admin.ID, Description: "创建管理员 " + admin.Username, CreatedAt: time.Now()})
	return adminManagerResponse(admin), nil
}

func (s *AdminService) UpdateAdmin(operatorID, id int64, req *dto.UpdateAdminManagerRequest) (*dto.AdminManagerResponse, error) {
	admin, err := s.adminRepo.FindAdminByID(id)
	if err != nil { return nil, errors.New(errors.ErrCodeNotFound, "管理员不存在") }
	if admin.Role == "super_admin" { return nil, errors.New(errors.ErrCodeForbidden, "超级管理员不能在此修改") }
	updates := map[string]interface{}{"name": req.Name, "role": req.Role}
	if req.IsActive != nil { updates["is_active"] = *req.IsActive }
	if req.Password != "" { hash, hashErr := utils.HashPassword(req.Password); if hashErr != nil { return nil, hashErr }; updates["password_hash"] = hash }
	if err := s.adminRepo.UpdateAdmin(id, updates); err != nil { return nil, err }
	updated, err := s.adminRepo.FindAdminByID(id)
	if err != nil { return nil, err }
	s.recordAudit(&model.AdminLog{AdminID: operatorID, Action: "update_admin", TargetType: "admin", TargetID: id, Description: "更新管理员 " + admin.Username, CreatedAt: time.Now()})
	return adminManagerResponse(updated), nil
}
