package repository

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// FindByUsername 根据用户名查找管理员
func (r *AdminRepository) FindByUsername(username string) (*model.AdminUser, error) {
	var admin model.AdminUser
	err := r.db.Where("username = ? AND is_active = ?", username, true).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// UpdateLastLogin 更新最后登录时间
func (r *AdminRepository) UpdateLastLogin(id int64, ip string) error {
	now := time.Now()
	return r.db.Model(&model.AdminUser{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
		}).Error
}

// CreateLog 创建操作日志
func (r *AdminRepository) CreateLog(log *model.AdminLog) error {
	if log.BeforeData == "" {
		log.BeforeData = "{}"
	}
	if log.AfterData == "" {
		log.AfterData = "{}"
	}
	return r.db.Create(log).Error
}

// GetStats 获取统计数据
func (r *AdminRepository) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总用户数
	var totalUsers int64
	r.db.Model(&model.User{}).Count(&totalUsers)
	stats["total_users"] = totalUsers

	// 总任务数
	var totalTasks int64
	r.db.Model(&model.Task{}).Count(&totalTasks)
	stats["total_tasks"] = totalTasks

	// 总作品数
	var totalWorks int64
	r.db.Model(&model.Work{}).Count(&totalWorks)
	stats["total_works"] = totalWorks

	// 待审核作品数
	var pendingWorks int64
	r.db.Model(&model.Work{}).Where("audit_status = ?", "pending").Count(&pendingWorks)
	stats["pending_works"] = pendingWorks

	// 今日新增用户
	today := time.Now().Format("2006-01-02")
	var todayUsers int64
	r.db.Model(&model.User{}).Where("DATE(created_at) = ?", today).Count(&todayUsers)
	stats["today_users"] = todayUsers

	// 今日新增任务
	var todayTasks int64
	r.db.Model(&model.Task{}).Where("DATE(created_at) = ?", today).Count(&todayTasks)
	stats["today_tasks"] = todayTasks

	// 今日新增作品
	var todayWorks int64
	r.db.Model(&model.Work{}).Where("DATE(created_at) = ?", today).Count(&todayWorks)
	stats["today_works"] = todayWorks

	return stats, nil
}
