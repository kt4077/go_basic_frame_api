package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"server_api/internal/common/plugininstaller"
)

// pluginCmd 提供插件包校验、安装、升级和静态注册代码生成命令。
func pluginCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "plugin", Short: "管理源码插件发行包"}
	cmd.AddCommand(
		pluginValidateCmd(), pluginGenerateCmd(), pluginInstallCmd(false), pluginInstallCmd(true),
		pluginRemoveCmd(), pluginPurgeCmd(),
	)
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
			if !result.DatabaseApplied {
				fmt.Println("当前仅安装源码，尚未执行数据库迁移、菜单和插件记录；确认配置后重新执行本命令并增加 --apply-database")
			} else if !upgrade {
				fmt.Println("插件已按安全策略以停用状态安装；请在插件管理中启用，并在启用后重启管理端API和用户端API，否则插件路由不会注册并返回404")
			}
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

func pluginRemoveCmd() *cobra.Command {
	var options plugininstaller.RemoveOptions
	cmd := &cobra.Command{
		Use:   "remove <plugin_id>",
		Short: "非破坏性移除插件，保留业务表、迁移历史和安装日志",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := plugininstaller.Remove(args[0], options)
			if err != nil {
				return err
			}
			fmt.Printf("插件%s移除完成，菜单=%d，后端源码=%t，管理端源码=%t；业务表、迁移历史、安装日志和上传文件已保留\n",
				result.PluginID, result.MenusRemoved, result.ServerRemoved, result.AdminRemoved)
			return nil
		},
	}
	bindPluginRemoveFlags(cmd, &options)
	return cmd
}

func pluginPurgeCmd() *cobra.Command {
	var options plugininstaller.RemoveOptions
	var confirmation string
	cmd := &cobra.Command{
		Use:   "purge <plugin_id>",
		Short: "彻底清理插件源码、业务表和插件记录",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if confirmation != args[0] {
				return fmt.Errorf("彻底清理不可恢复，请使用 --confirm %s 明确确认", args[0])
			}
			result, err := plugininstaller.Purge(args[0], options)
			if err != nil {
				return err
			}
			fmt.Printf("插件%s彻底清理完成，删除业务表=%v；上传文件未自动删除\n", result.PluginID, result.DroppedTables)
			return nil
		},
	}
	bindPluginRemoveFlags(cmd, &options)
	cmd.Flags().StringVar(&confirmation, "confirm", "", "必须填写与plugin_id完全一致的确认值")
	return cmd
}

func bindPluginRemoveFlags(cmd *cobra.Command, options *plugininstaller.RemoveOptions) {
	cmd.Flags().StringVar(&options.ServerRoot, "server-root", ".", "server_api根目录")
	cmd.Flags().StringVar(&options.AdminRoot, "admin-root", "../admin_client", "admin_client根目录")
	cmd.Flags().StringVarP(&options.ConfigPath, "config", "c", "config.yaml", "数据库配置文件")
}
