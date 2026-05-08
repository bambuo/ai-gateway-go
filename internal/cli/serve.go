package cli

import (
	"fmt"
	"os"

	"ai/gateway/internal/auth"
	"ai/gateway/internal/config"
	"ai/gateway/internal/database"
	"ai/gateway/internal/logger"
	"ai/gateway/internal/oauth"
	"ai/gateway/internal/proxy"
	"ai/gateway/internal/rewriter"

	"github.com/spf13/cobra"
)

var serveDBPath string

var serveCmd = &cobra.Command{
	Use:     "serve [config-path]",
	Aliases: []string{"start"},
	Short:   "启动 AI Gateway 代理服务器",
	Long: `使用指定的配置文件或数据库启动 AI Gateway 代理服务器。

支持两种配置来源（优先级：DB > 文件）：
  1. --db ./data/admin.db  从数据库加载完整配置（推荐）
  2. [config-path]          从 config.yaml 加载配置（传统方式）

如果省略 config-path，默认使用 "config.yaml"。`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var cfg *config.Config

		if serveDBPath != "" {
			logger.Info("正在从数据库加载配置...", "db", serveDBPath)
			if err := database.Init(serveDBPath); err != nil {
				return fmt.Errorf("初始化数据库: %w", err)
			}
			if !database.HasAppConfig() {
				return fmt.Errorf("数据库中无配置，请先通过管理后台初始化系统 (gateway admin)")
			}
			fc, err := database.GetActiveFullConfig()
			if err != nil {
				return fmt.Errorf("从数据库加载配置: %w", err)
			}
			cfg = config.FullConfigToAppConfig(fc)
		} else {
			configPath := "config.yaml"
			if len(args) > 0 {
				configPath = args[0]
			}
			var err error
			cfg, err = config.Load(configPath)
			if err != nil {
				return fmt.Errorf("加载配置: %w", err)
			}
		}

		logger.Setup(cfg.Logging.Level)

		tokMgr := oauth.New(cfg.OAuth)
		tokMgr.Init()
		defer tokMgr.Stop()

		authenticator := auth.New(cfg.Auth)
		rw := rewriter.New(cfg)

		srv := proxy.NewServer(cfg, authenticator, tokMgr, rw, rw)

		if serveDBPath != "" {
			logger.Info("AI Gateway 已启动（数据库模式）", "config_source", serveDBPath)
		} else {
			configPath := "config.yaml"
			if len(args) > 0 {
				configPath = args[0]
			}
			logger.Info("AI Gateway 已启动（文件模式）", "config", configPath)
		}

		if err := srv.Start(); err != nil {
			logger.Error("服务器错误", "error", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	serveCmd.Flags().StringVarP(&serveDBPath, "db", "d", "", "从 SQLite 数据库加载配置（替代 config.yaml）")
	rootCmd.AddCommand(serveCmd)
}
