package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"server_api/config"
	"server_api/internal/common/model"
)

// App 全局运行时依赖，由各子命令启动时初始化。
type App struct {
	Cfg   *config.Config
	DB    *gorm.DB
	Redis *redis.Client
}

// Init 初始化数据库与 Redis 连接。
func Init(cfg *config.Config) (*App, error) {
	db, err := gorm.Open(mysql.Open(cfg.Mysql.DSN()), &gorm.Config{
		Logger: gormlogger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormlogger.Config{
			SlowThreshold: time.Second,
			LogLevel:      gormlogger.Warn,
		}),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.Mysql.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Mysql.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	mysqlCtx, mysqlCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := sqlDB.PingContext(mysqlCtx); err != nil {
		mysqlCancel()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("MySQL 健康检查失败: %w", err)
	}
	mysqlCancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	redisCtx, redisCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer redisCancel()
	if err := rdb.Ping(redisCtx).Err(); err != nil {
		_ = rdb.Close()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("Redis 健康检查失败: %w", err)
	}

	return &App{Cfg: cfg, DB: db, Redis: rdb}, nil
}

// Close 释放应用持有的连接池。可重复调用方应只在服务生命周期结束时调用。
func (a *App) Close() error {
	var firstErr error
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			firstErr = err
		}
	}
	if a.DB != nil {
		sqlDB, err := a.DB.DB()
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
		} else if err := sqlDB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// CheckTables 启动前检查必需的数据表是否存在，缺失时直接报错并提示初始化方式。
// 本项目不使用自动迁移，建表与初始数据由 sql/schema.sql 手动执行。
func (a *App) CheckTables() error {
	required := []string{
		"sys_dept", "sys_user", "sys_role", "sys_menu",
		"sys_user_login", "sys_user_role", "sys_role_menu",
		"sys_operation_log", "sys_storage_config", "sys_upload_file",
		"sys_sms_config", "sys_sms_signature", "sys_sms_template", "sys_sms_send_log",
		"sys_wechat_config", "sys_payment_config",
	}
	var missing []string
	for _, table := range required {
		if !a.DB.Migrator().HasTable(table) {
			missing = append(missing, table)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少数据表 %v，请先执行 sql/schema.sql 初始化数据库", missing)
	}
	var userCount int64
	a.DB.Model(&model.SysUser{}).Count(&userCount)
	if userCount == 0 {
		return fmt.Errorf("sys_user 表没有数据，请执行 sql/schema.sql 导入初始账号")
	}
	return nil
}
