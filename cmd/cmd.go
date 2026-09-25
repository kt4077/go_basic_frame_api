// Package cmd 子命令实现。
// 启动方式：server_api service admin / server_api service api
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"server_api/config"
	commonapp "server_api/internal/common/app"
	commonplugin "server_api/internal/common/plugin"
	"server_api/internal/plugins"
	"server_api/router"
)

// RootCmd 根命令。
func RootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "server_api",
		Short: "Go 后端基础框架服务",
	}
	root.AddCommand(serviceCmd(), versionCmd(), pluginCmd())
	return root
}

// serviceCmd 服务启动命令，二级子命令区分端口角色。
func serviceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "启动服务",
		Long: `启动后端服务，通过二级子命令区分角色：
  server_api service admin   启动管理端 API（默认 :8001）
  server_api service api     启动用户端 API（默认 :8002）`,
	}
	cmd.AddCommand(adminCmd(), apiCmd())
	return cmd
}

// adminCmd 管理端 API 服务：service admin [-c config.yaml]
func adminCmd() *cobra.Command {
	var cfgPath string
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "启动管理端 API 服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(cfgPath, "admin")
		},
	}
	cmd.Flags().StringVarP(&cfgPath, "config", "c", "config.yaml", "配置文件路径")
	return cmd
}

// apiCmd 用户端 API 服务：service api [-c config.yaml]
func apiCmd() *cobra.Command {
	var cfgPath string
	cmd := &cobra.Command{
		Use:   "api",
		Short: "启动用户端 API 服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(cfgPath, "api")
		},
	}
	cmd.Flags().StringVarP(&cfgPath, "config", "c", "config.yaml", "配置文件路径")
	return cmd
}

func versionCmd() *cobra.Command {
	var cfgPath string
	cmd := &cobra.Command{
		Use:   "version",
		Short: "查看版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}
			fmt.Println("server_api " + cfg.Version)
			return nil
		},
	}
	cmd.Flags().StringVarP(&cfgPath, "config", "c", "config.yaml", "配置文件路径")
	return cmd
}

func serve(cfgPath, mode string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	app, err := commonapp.Init(cfg)
	if err != nil {
		return err
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("关闭依赖连接失败: %v", err)
		}
	}()
	// 启动前检查数据表，缺表/无初始数据时给出明确提示，避免登录时报错误导排查
	if err := app.CheckTables(); err != nil {
		return err
	}

	pluginRegistry := commonplugin.NewRegistry(app)
	if err := plugins.RegisterBuiltins(pluginRegistry); err != nil {
		return fmt.Errorf("注册内置插件失败: %w", err)
	}
	if err := pluginRegistry.Prepare(context.Background()); err != nil {
		return fmt.Errorf("插件检查失败: %w", err)
	}

	addr := cfg.Server.ApiAddr
	serviceType := commonplugin.ServiceAPI
	engine, err := router.ApiRoutes(app, pluginRegistry)
	serviceName := "用户端 API"
	if mode == "admin" {
		addr = cfg.Server.AdminAddr
		serviceType = commonplugin.ServiceAdmin
		engine, err = router.AdminRoutes(app, pluginRegistry)
		serviceName = "管理端 API"
	}
	if err != nil {
		return fmt.Errorf("注册%s路由失败: %w", serviceName, err)
	}

	pluginCtx, pluginCancel := context.WithCancel(context.Background())
	if err := pluginRegistry.Start(pluginCtx, serviceType); err != nil {
		pluginCancel()
		return err
	}
	defer func() {
		pluginCancel()
		stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
		defer stopCancel()
		if err := pluginRegistry.Stop(stopCtx); err != nil {
			log.Printf("停止插件失败: %v", err)
		}
	}()

	server := &http.Server{
		Addr:              addr,
		Handler:           http.Handler(engine),
		ReadHeaderTimeout: time.Duration(cfg.Server.ReadHeaderTimeout) * time.Second,
		ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("%s %s 启动: http://0.0.0.0%s\n", serviceName, cfg.Version, addr)
		errCh <- server.ListenAndServe()
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-signalCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("服务优雅停机失败: %w", err)
		}
		return nil
	}
}
