package cmd

import (
	"fmt"

	"github.com/fluxrelay/flux-relay-cli/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from Flux Relay",
	Long: `Log out from Flux Relay by removing the stored access token.
You will need to run 'flux-relay login' again to authenticate.`,
	RunE: runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	cfg := config.New()

	// Check if token exists
	token, err := cfg.GetToken()
	if err != nil || token == nil {
		fmt.Println("ℹ️  No active session found. You are already logged out.")
		return nil
	}

	// Remove token
	if err := cfg.RemoveToken(); err != nil {
		return fmt.Errorf("failed to remove token: %w", err)
	}

	fmt.Println("✅ Logged out successfully")
	fmt.Println("   Token removed from:", cfg.ConfigPath())

	return nil
}
