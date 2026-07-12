package service

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/config"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/jwt"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"gorm.io/gorm"
)

type AdminService struct {
	adminRepo *repository.AdminRepository
	workRepo  *repository.WorkRepository
	userRepo  *repository.UserRepository
	jwtConfig *config.JWTConfig
}

func NewAdminService(
	adminRepo *repository.AdminRepository,
	workRepo *repository.WorkRepository,
	userRepo *repository.UserRepository,
	jwtConfig *config.JWTConfig,
) *AdminService {
	return &AdminService{
		adminRepo: adminRepo,
		workRepo:  workRepo,
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

// Login 管理员登录
func (s *AdminService) Login(req *dto.AdminLoginRequest, ip string) (*dto.AdminLoginResponse, error) {
	// 1. 查找管理员
	admin, err := s.adminRepo.FindByUsername(req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeUnauthorized, "用户名或密码错误")
		}
		return nil, err
	}

	// 2. 验证密码
	if !utils.CheckPassword(req.Password, admin.PasswordHash) {
		return nil, errors.New(errors.ErrCodeUnauthorized, "用户名或密码错误")
	}

	// 3. 更新登录信息
	if err := s.adminRepo.UpdateLastLogin(admin.ID, ip); err != nil {
		return nil, err
	}

	// 4. 生成Token（管理员）
	token, err := jwt.GenerateAdminToken(admin.ID, admin.Username, admin.Name, s.jwtConfig)
	if err != nil {
		return nil, err
	}

	// 5. 记录登录日志
	if err := s.adminRepo.CreateLog(&model.AdminLog{
		AdminID:     admin.ID,
		Action:      "login",
		Description: "管理员登录",
		IP:          ip,
		CreatedAt:   time.Now(),
	}); err != nil {
		return nil, err
	}

	return &dto.AdminLoginResponse{
		Token: token,
		Admin: &dto.AdminInfo{
			ID:       admin.ID,
			Username: admin.Username,
			Name:     admin.Name,
			Role:     admin.Role,
		},
	}, nil
}

// AuditWork 审核作品
func (s *AdminService) AuditWork(adminID, workID int64, req *dto.AdminWorkAuditRequest) error {
	// 1. 获取作品
	work, err := s.workRepo.FindByID(workID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrWorkNotFound
		}
		return err
	}

	// 2. 检查审核状态
	if work.AuditStatus != "pending" {
		return errors.New(errors.ErrCodeBadRequest, "该作品已审核")
	}

	// 3. 更新作品状态
	now := time.Now()
	work.AuditStatus = req.Status
	work.AuditReason = req.Reason
	work.AuditAdminID = adminID
	work.AuditedAt = &now
	work.UpdatedAt = now

	if err := s.workRepo.Update(work); err != nil {
		return err
	}

	// 4. 创建审核记录
	audit := &model.WorkAudit{
		WorkID:     workID,
		AdminID:    adminID,
		Status:     req.Status,
		Reason:     req.Reason,
		AuditedAt:  now,
		CreatedAt:  now,
	}
	if err := s.workRepo.CreateAudit(audit); err != nil {
		return err
	}

	// 5. 记录操作日志
	if err := s.adminRepo.CreateLog(&model.AdminLog{
		AdminID:     adminID,
		Action:      "audit_work",
		TargetType:  "work",
		TargetID:    workID,
		Description: "审核作品：" + req.Status,
		CreatedAt:   time.Now(),
	}); err != nil {
		return err
	}

	return nil
}

// GetPendingWorks 获取待审核作品列表
func (s *AdminService) GetPendingWorks(page, pageSize int) ([]*dto.WorkResponse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	works, total, err := s.workRepo.FindPendingWorks(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.WorkResponse, len(works))
	for i, work := range works {
		responses[i] = &dto.WorkResponse{
			ID:          work.ID,
			UserID:      work.UserID,
			Title:       work.Title,
			Prompt:      work.Prompt,
			Category:    work.Category,
			Type:        work.Type,
			ImageURL:    work.ImageURL,
			VideoURL:    work.VideoURL,
			AuditStatus: work.AuditStatus,
			CreatedAt:   work.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return responses, total, nil
}

// GetStats 获取统计数据
func (s *AdminService) GetStats() (*dto.AdminStatsResponse, error) {
	stats, err := s.adminRepo.GetStats()
	if err != nil {
		return nil, err
	}

	return &dto.AdminStatsResponse{
		TotalUsers:   stats["total_users"],
		TotalTasks:   stats["total_tasks"],
		TotalWorks:   stats["total_works"],
		PendingWorks: stats["pending_works"],
		TodayUsers:   stats["today_users"],
		TodayTasks:   stats["today_tasks"],
		TodayWorks:   stats["today_works"],
	}, nil
}
