package cli

import (
	"context"
	"fmt"
	"log"
	"os"

	"ai/gateway/internal/auth"
	"ai/gateway/internal/config"
	"ai/gateway/internal/logger"
	"ai/gateway/internal/oauth"
	"ai/gateway/internal/proxy"
	"ai/gateway/internal/rewriter"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:     "serve [config-path]",
	Aliases: []string{"start"},
	Short: "启动 AI Gateway 代理服务器",
	Long: `使用指定的配置文件启动 AI Gateway 代理服务器。
如果省略 config-path，默认使用 "config.yaml"。`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := "config.yaml"
		if len(args) > 0 {
			configPath = args[0]
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("加载配置: %w", err)
		}

		logger.Setup(cfg.Logging.Level)

		tokMgr := oauth.New(cfg.OAuth)
		if err := tokMgr.Init(context.Background()); err != nil {
			return fmt.Errorf("初始化 OAuth: %w", err)
		}
		defer tokMgr.Stop()

		authenticator := auth.New(cfg.Auth)
		rw := rewriter.New(cfg)

		srv := proxy.NewServer(cfg, authenticator, tokMgr, rw, rw)

		fmt.Fprintf(os.Stderr, "AI Gateway 已启动 — 配置文件: %s\n", configPath)
		if err := srv.Start(); err != nil {
			log.Fatalf("服务器错误: %v", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
