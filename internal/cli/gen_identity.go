package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
)

var genIdentityCmd = &cobra.Command{
	Use:   "gen-identity",
	Short: "Generate a canonical device identity",
	Long: `Generate a random 32-byte device identity that can be used
in the config.yaml under the identity section.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("generate random bytes: %w", err)
		}
		deviceID := hex.EncodeToString(b)

		fmt.Println("\nGenerated canonical identity:")
		fmt.Println()
		fmt.Println("identity:")
		fmt.Printf("  device_id: \"%s\"\n", deviceID)
		fmt.Println()
		fmt.Println("Put this in your config.yaml. All clients will appear as this device.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genIdentityCmd)
}
