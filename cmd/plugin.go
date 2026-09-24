package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"server_api/internal/common/plugininstaller"
)

// pluginCmd 提供插件包校验、安装、升级和静态注册代码生成命令。
func pluginCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "plugin", Short: "管理源码插件发行包"}
	cmd.AddCommand(pluginValidateCmd(), pluginGenerateCmd(), pluginInstallCmd(false), pluginInstallCmd(true))
	return cmd
}

func pluginValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <package.zip>",
		Short: "校验插件包清单、路径、摘要、菜单和迁移SQL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg, err := plugininstaller.OpenPackage(args[0])
			if err != nil {
				return err
			}
			defer pkg.Close()
			fmt.Printf("插件包校验通过: %s %s, SHA256=%s\n", pkg.Manifest.PluginID, pkg.Manifest.Version, pkg.Hash)
			return nil
		},
	}
}

func pluginGenerateCmd() *cobra.Command {
	var serverRoot string
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "扫描插件目录并生成静态注册文件",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := plugininstaller.GenerateRegistry(serverRoot); err != nil {
				return err
			}
			fmt.Println("插件注册文件生成完成")
			return nil
		},
	}
	cmd.Flags().StringVar(&serverRoot, "server-root", ".", "server_api根目录")
	return cmd
}

func pluginInstallCmd(upgrade bool) *cobra.Command {
	var options plugininstaller.InstallOptions
	use := "install <package.zip>"
	short := "安装插件源码并自动生成注册文件"
	if upgrade {
		use = "upgrade <package.zip>"
		short = "升级已安装插件源码并自动生成注册文件"
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg, err := plugininstaller.OpenPackage(args[0])
			if err != nil {
				return err
			}
			defer pkg.Close()
			options.Upgrade = upgrade
			result, err := plugininstaller.Install(pkg, options)
			if err != nil {
				return err
			}
			fmt.Printf("插件%s %s处理完成，后端=%t，管理端=%t，数据库=%t\n",
				result.PluginID, result.Version, result.ServerInstalled, result.AdminInstalled, result.DatabaseApplied)
			fmt.Println("请完成代码检查和构建；启用或升级插件后重启管理端API与用户端API")
			return nil
		},
	}
	cmd.Flags().StringVar(&options.ServerRoot, "server-root", ".", "server_api根目录")
	cmd.Flags().StringVar(&options.AdminRoot, "admin-root", "../admin_client", "admin_client根目录")
	cmd.Flags().StringVarP(&options.ConfigPath, "config", "c", "config.yaml", "数据库配置文件")
	cmd.Flags().BoolVar(&options.ApplyDatabase, "apply-database", false, "校验后执行插件迁移、菜单和安装记录")
	return cmd
}
