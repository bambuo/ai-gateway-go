package cli

import (
	"fmt"

	"ai/gateway/internal/admin"
	"ai/gateway/internal/database"
	"ai/gateway/internal/logger"

	"github.com/spf13/cobra"
)

var (
	adminPort   int
	adminDBPath string
	adminStatic string
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "启动 Web 管理后台",
	Long: `启动 AI Gateway Web 管理后台服务器。
提供浏览器界面进行系统配置管理：
- 系统初始化（网关配置 + 管理员账户创建）
- 管理员登录认证
- 仪表盘监控`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Setup("info")
		logger.Info("正在初始化管理后台数据库...", "path", adminDBPath)
		if err := database.Init(adminDBPath); err != nil {
			return fmt.Errorf("数据库初始化失败: %w", err)
		}
		srv := admin.NewServer(adminPort, adminStatic)
		logger.Info("管理后台服务已启动", "port", adminPort, "database", adminDBPath, "static", adminStatic)
		if err := srv.Start(); err != nil {
			return fmt.Errorf("管理后台服务错误: %w", err)
		}
		return nil
	},
}

func init() {
	adminCmd.Flags().IntVarP(&adminPort, "port", "p", 8080, "管理后台监听端口")
	adminCmd.Flags().StringVarP(&adminDBPath, "db", "d", "./data/admin.db", "SQLite 数据库文件路径")
	adminCmd.Flags().StringVarP(&adminStatic, "static", "s", "./web/dist", "前端静态文件目录路径")
	rootCmd.AddCommand(adminCmd)
}
