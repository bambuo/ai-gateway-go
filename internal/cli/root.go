package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gateway",
	Short: "AI Gateway — AI API 身份网关与请求重写代理",
	Long: `AI Gateway 是一个 AI 服务反向代理服务器，提供以下功能：
- 基于令牌的身份认证
- OAuth2 客户端凭证管理
- 请求/响应内容重写
- SSE 事件流转换`,
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
