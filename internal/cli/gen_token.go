package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
)

var genTokenCmd = &cobra.Command{
	Use:   "gen-token [name]",
	Short: "生成认证令牌",
	Long: `为客户端生成一个 32 字节的随机认证令牌。
可选的 name 参数指定客户端名称（默认："client-1"）。`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := make([]byte, 32)
		if _, err := rand.Read(token); err != nil {
			return fmt.Errorf("生成随机令牌: %w", err)
		}

		name := "client-1"
		if len(args) > 0 {
			name = args[0]
		}

		fmt.Println("\n请将以下内容添加到 config.yaml 的 auth.tokens 下：")
		fmt.Println()
		fmt.Printf("  - name: %s\n", name)
		fmt.Printf("    token: %s\n", hex.EncodeToString(token))
		fmt.Println()
		fmt.Println("客户端应设置请求头：")
		fmt.Printf("  Authorization: Bearer %s\n", hex.EncodeToString(token))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genTokenCmd)
}
