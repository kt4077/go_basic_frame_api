// Package config 负责加载配置文件。
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string       `yaml:"version"`
	Server  ServerConfig `yaml:"server"`
	Mysql   MysqlConfig  `yaml:"mysql"`
	Redis   RedisConfig  `yaml:"redis"`
	Jwt     JwtConfig    `yaml:"jwt"`
	Log     LogConfig    `yaml:"log"`
}

type ServerConfig struct {
	AdminAddr         string `yaml:"admin_addr"`
	ApiAddr           string `yaml:"api_addr"`
	ReadHeaderTimeout int    `yaml:"read_header_timeout_seconds"`
	ReadTimeout       int    `yaml:"read_timeout_seconds"`
	WriteTimeout      int    `yaml:"write_timeout_seconds"`
	IdleTimeout       int    `yaml:"idle_timeout_seconds"`
	ShutdownTimeout   int    `yaml:"shutdown_timeout_seconds"`
}

type MysqlConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

func (c MysqlConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JwtConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
	Issuer      string `yaml:"issuer"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

// Load 从指定路径加载配置文件。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	setDefaults(cfg)
	if cfg.Version == "" {
		return nil, fmt.Errorf("version 不能为空，请在配置文件中设置系统版本号")
	}
	if cfg.Server.AdminAddr == "" || cfg.Server.ApiAddr == "" {
		return nil, fmt.Errorf("server.admin_addr 和 server.api_addr 不能为空")
	}
	if cfg.Mysql.MaxOpenConns <= 0 || cfg.Mysql.MaxIdleConns < 0 || cfg.Mysql.MaxIdleConns > cfg.Mysql.MaxOpenConns {
		return nil, fmt.Errorf("MySQL 连接池配置无效")
	}
	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.ReadHeaderTimeout <= 0 {
		cfg.Server.ReadHeaderTimeout = 5
	}
	if cfg.Server.ReadTimeout <= 0 {
		cfg.Server.ReadTimeout = 60
	}
	if cfg.Server.WriteTimeout <= 0 {
		cfg.Server.WriteTimeout = 60
	}
	if cfg.Server.IdleTimeout <= 0 {
		cfg.Server.IdleTimeout = 120
	}
	if cfg.Server.ShutdownTimeout <= 0 {
		cfg.Server.ShutdownTimeout = 15
	}
}
