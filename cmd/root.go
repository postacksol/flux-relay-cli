package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	apiBaseURL  string
	verbose     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "flux-relay",
	Short: "Flux Relay CLI - Manage your messaging platform from the command line",
	Long: `Flux Relay CLI is a command-line tool for managing your Flux Relay
messaging platform. Execute SQL queries, manage namespaces, and more.`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.flux-relay/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiBaseURL, "api-url", "", "API base URL (default: https://flux.postacksolutions.com)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("api_url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".flux-relay" (without extension).
		viper.AddConfigPath(home + "/.flux-relay")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// getAPIURL returns the API URL from flag, config, or default production URL
func getAPIURL() string {
	if apiBaseURL != "" {
		return apiBaseURL
	}
	if url := viper.GetString("api_url"); url != "" {
		return url
	}
	// Default to production URL
	return "https://flux.postacksolutions.com"
}