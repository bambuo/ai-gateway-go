package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
)

var genIdentityCmd = &cobra.Command{
	Use:   "gen-identity",
	Short: "生成设备身份标识",
	Long: `生成一个 32 字节的随机设备身份标识，
可用于 config.yaml 中的 identity 配置段。`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("生成随机字节: %w", err)
		}
		deviceID := hex.EncodeToString(b)

		fmt.Println("\n已生建设备身份标识：")
		fmt.Println()
		fmt.Println("identity:")
		fmt.Printf("  device_id: \"%s\"\n", deviceID)
		fmt.Println()
		fmt.Println("请将此配置添加到 config.yaml 中。所有客户端将以该设备身份出现。")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genIdentityCmd)
}
