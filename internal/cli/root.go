package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gateway",
	Short: "AI Gateway — AI proxy server with OAuth & request rewriting",
	Long: `AI Gateway is a reverse proxy server for AI services that provides:
- Token-based authentication
- OAuth2 client credentials management
- Request/response rewriting
- SSE event transformation`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
