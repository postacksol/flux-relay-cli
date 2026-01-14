package cmd

import (
	"fmt"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  "Manage CLI configuration settings and tokens",
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration value",
	Long:  "Set configuration values like token",
}

var configSetTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Set the access token",
	Long:  "Set the access token for authentication",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigSetToken,
}

func init() {
	configSetCmd.AddCommand(configSetTokenCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigSetToken(cmd *cobra.Command, args []string) error {
	token := args[0]

	// Get API URL
	apiURL := apiBaseURL
	if apiURL == "" {
		apiURL = viper.GetString("api_url")
		if apiURL == "" {
			apiURL = "http://localhost:3000"
		}
	}

	// Validate token by getting user info
	client := api.NewClient(apiURL)
	userInfo, err := client.GetCurrentUser(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Create token response
	tokenResponse := &api.TokenResponse{
		AccessToken:  token,
		RefreshToken: "", // Not available when setting manually
		TokenType:    "Bearer",
		ExpiresIn:    86400, // Default 24 hours
		Developer: struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}{
			ID:    userInfo.ID(),
			Email: userInfo.Email(),
		},
	}

	// Save token
	cfg := config.New()
	if err := cfg.SaveToken(tokenResponse); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("Token saved successfully!")
	fmt.Printf("   Logged in as: %s (%s)\n", userInfo.Email(), userInfo.ID())
	fmt.Printf("   Token saved to: %s\n", cfg.ConfigPath())
	fmt.Println()

	return nil
}
