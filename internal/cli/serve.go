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
	Short:   "Start the AI gateway proxy server",
	Long: `Start the AI gateway proxy server with the given configuration file.
If config-path is omitted, it defaults to "config.yaml".`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := "config.yaml"
		if len(args) > 0 {
			configPath = args[0]
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		logger.Setup(cfg.Logging.Level)

		tokMgr := oauth.New(cfg.OAuth)
		if err := tokMgr.Init(context.Background()); err != nil {
			return fmt.Errorf("init oauth: %w", err)
		}
		defer tokMgr.Stop()

		authenticator := auth.New(cfg.Auth)
		rw := rewriter.New(cfg)

		srv := proxy.NewServer(cfg, authenticator, tokMgr, rw, rw)

		fmt.Fprintf(os.Stderr, "AI Gateway started — config: %s\n", configPath)
		if err := srv.Start(); err != nil {
			log.Fatalf("server: %v", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
