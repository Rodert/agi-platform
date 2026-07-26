package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

type ReportRecords struct {
	Users   []*model.User
	Tasks   []*model.Task
	Ledgers []*model.CreditLedger
	Works   []*model.Work
}

type DatabaseTable struct {
	Name    string
	Comment string
}

type DatabaseColumn struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// ListDatabaseTables discovers base tables from the current schema, so database
// migrations automatically appear in the admin data browser.
func (r *AdminRepository) ListDatabaseTables() ([]DatabaseTable, error) {
	var tables []DatabaseTable
	err := r.db.Raw(`SELECT table_name AS name, table_comment AS comment
		FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'
		ORDER BY table_name`).Scan(&tables).Error
	return tables, err
}

func (r *AdminRepository) GetDatabaseTable(tableName string, page, pageSize int) ([]DatabaseColumn, []map[string]interface{}, int64, error) {
	tables, err := r.ListDatabaseTables()
	if err != nil { return nil, nil, 0, err }
	allowed := false
	for _, table := range tables {
		if table.Name == tableName { allowed = true; break }
	}
	if !allowed { return nil, nil, 0, gorm.ErrRecordNotFound }

	var columns []DatabaseColumn
	err = r.db.Raw(`SELECT column_name AS name, column_type AS type,
		CASE WHEN is_nullable = 'YES' THEN 1 ELSE 0 END AS nullable,
		CASE WHEN column_key = 'PRI' THEN 1 ELSE 0 END AS primary_key
		FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ?
		ORDER BY ordinal_position`, tableName).Scan(&columns).Error
	if err != nil { return nil, nil, 0, err }

	quotedTable := "`" + strings.ReplaceAll(tableName, "`", "``") + "`"
	var total int64
	if err := r.db.Raw("SELECT COUNT(*) FROM " + quotedTable).Scan(&total).Error; err != nil { return nil, nil, 0, err }

	orderBy := ""
	for _, column := range columns {
		if column.PrimaryKey {
			orderBy = " ORDER BY `" + strings.ReplaceAll(column.Name, "`", "``") + "` DESC"
			break
		}
	}
	rows := make([]map[string]interface{}, 0)
	query := fmt.Sprintf("SELECT * FROM %s%s LIMIT ? OFFSET ?", quotedTable, orderBy)
	if err := r.db.Raw(query, pageSize, (page-1)*pageSize).Scan(&rows).Error; err != nil { return nil, nil, 0, err }
	return columns, rows, total, nil
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

func (r *AdminRepository) FindAdminByUsername(username string) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.Where("username = ?", username).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) FindByID(id int64) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.Where("id = ? AND is_active = ?", id, true).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) FindAdminByID(id int64) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.Where("id = ?", id).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) ListAdmins() ([]*model.AdminUser, error) {
	var admins []*model.AdminUser
	err := r.db.Order("id ASC").Find(&admins).Error
	return admins, err
}

func (r *AdminRepository) CreateAdmin(admin *model.AdminUser) error {
	return r.db.Create(admin).Error
}

func (r *AdminRepository) UpdateAdmin(id int64, updates map[string]interface{}) error {
	return r.db.Model(&model.AdminUser{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AdminRepository) UpdateProfile(id int64, name, passwordHash string) error {
	updates := map[string]interface{}{"name": name}
	if passwordHash != "" {
		updates["password_hash"] = passwordHash
	}
	return r.db.Model(&model.AdminUser{}).Where("id = ?", id).Updates(updates).Error
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

// ListLogs returns audit records with their actor, newest first.
func (r *AdminRepository) ListLogs(operator, action string, startAt, endAt *time.Time, loginOnly bool, page, pageSize int) ([]*model.AdminLog, int64, error) {
	var logs []*model.AdminLog
	var total int64
	query := r.db.Model(&model.AdminLog{}).Joins("LEFT JOIN admin_users ON admin_users.id = admin_logs.admin_id")
	if operator = strings.TrimSpace(operator); operator != "" {
		query = query.Where("admin_users.username LIKE ? OR admin_users.name LIKE ?", "%"+operator+"%", "%"+operator+"%")
	}
	if action = strings.TrimSpace(action); action != "" {
		query = query.Where("admin_logs.action = ?", action)
	} else if loginOnly {
		query = query.Where("admin_logs.action IN ?", []string{"login", "login_failed"})
	}
	if startAt != nil {
		query = query.Where("admin_logs.created_at >= ?", *startAt)
	}
	if endAt != nil {
		query = query.Where("admin_logs.created_at <= ?", *endAt)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Select("admin_logs.*").Preload("Admin").Order("admin_logs.created_at DESC, admin_logs.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
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

// GetReportRecords loads the bounded source records. Aggregation stays in the
// service layer so this repository contains no report-specific SQL.
func (r *AdminRepository) GetReportRecords(start, end time.Time) (*ReportRecords, error) {
	records := &ReportRecords{}
	if err := r.db.Where("created_at >= ? AND created_at < ?", start, end).Find(&records.Users).Error; err != nil {
		return nil, err
	}
	if err := r.db.Preload("Channel").Where("created_at >= ? AND created_at < ?", start, end).Find(&records.Tasks).Error; err != nil {
		return nil, err
	}
	if err := r.db.Where("type = ? AND created_at >= ? AND created_at < ?", "expense", start, end).Find(&records.Ledgers).Error; err != nil {
		return nil, err
	}
	if err := r.db.Where("created_at >= ? AND created_at < ?", start, end).Find(&records.Works).Error; err != nil {
		return nil, err
	}
	return records, nil
}
