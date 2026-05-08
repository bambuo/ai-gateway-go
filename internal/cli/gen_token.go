package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
)

var genTokenCmd = &cobra.Command{
	Use:   "gen-token [name]",
	Short: "Generate an authentication token",
	Long: `Generate a random 32-byte authentication token for a client.
The optional name argument specifies the client name (default: "client-1").`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := make([]byte, 32)
		if _, err := rand.Read(token); err != nil {
			return fmt.Errorf("generate random token: %w", err)
		}

		name := "client-1"
		if len(args) > 0 {
			name = args[0]
		}

		fmt.Println("\nAdd this to your config.yaml under auth.tokens:")
		fmt.Println()
		fmt.Printf("  - name: %s\n", name)
		fmt.Printf("    token: %s\n", hex.EncodeToString(token))
		fmt.Println()
		fmt.Println("Client should set:")
		fmt.Printf("  Authorization: Bearer %s\n", hex.EncodeToString(token))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genTokenCmd)
}
