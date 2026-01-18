package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
	"github.com/spf13/cobra"
)

var nsCmd = &cobra.Command{
	Use:   "ns [nameserver-name-or-id]",
	Short: "Manage nameservers",
	Long: `List and select nameservers (databases) in the selected server.

Examples:
  flux-relay ns list              # List all nameservers
  flux-relay ns db                # Select by name
  flux-relay ns db_123            # Select by ID
  flux-relay ns                   # Show current nameserver`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNsShowOrSelect,
}

var nsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all nameservers in the selected server",
	Long:  "List all nameservers (databases) in the currently selected server",
	RunE:  runNsList,
}

var nsShellCmd = &cobra.Command{
	Use:   "shell [nameserver-name-or-id]",
	Short: "Open interactive SQL shell for a nameserver",
	Long: `Open an interactive SQL shell for a nameserver, similar to Turso's shell.

Examples:
  flux-relay ns shell db
  flux-relay ns shell db_123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNameserverShell(args[0])
	},
}

func init() {
	nsCmd.AddCommand(nsListCmd)
	nsCmd.AddCommand(nsShellCmd)
	rootCmd.AddCommand(nsCmd)
}

func runNsShowOrSelect(cmd *cobra.Command, args []string) error {
	// Get API URL
	apiURL := getAPIURL()

	// Get access token
	cfg := config.New()
	accessToken := cfg.GetAccessToken()
	if accessToken == "" {
		return fmt.Errorf("not logged in. Run 'flux-relay login' first")
	}

	// Get selected project and server
	projectID := cfg.GetSelectedProject()
	if projectID == "" {
		return fmt.Errorf("no project selected. Use 'flux-relay pr <project-name-or-id>' to select a project")
	}

	serverID := cfg.GetSelectedServer()
	if serverID == "" {
		return fmt.Errorf("no server selected. Use 'flux-relay server <server-name-or-id>' to select a server")
	}

	// If no argument, show current nameserver
	if len(args) == 0 {
		selectedNameserverID := cfg.GetSelectedNameserver()
		if selectedNameserverID == "" {
			fmt.Println("No nameserver selected.")
			fmt.Println()
			fmt.Println("Select a nameserver using:")
			fmt.Println("  flux-relay ns <nameserver-name-or-id>")
			fmt.Println()
			fmt.Println("Or list available nameservers:")
			fmt.Println("  flux-relay ns list")
			return nil
		}

		// Get nameserver details
		client := api.NewClient(apiURL)
		databasesResponse, err := client.ListDatabases(accessToken, projectID, serverID)
		if err != nil {
			return fmt.Errorf("failed to get nameserver info: %w", err)
		}

		// Find the selected nameserver
		var selectedNameserver *api.Database
		for i := range databasesResponse.Databases {
			if databasesResponse.Databases[i].ID == selectedNameserverID {
				selectedNameserver = &databasesResponse.Databases[i]
				break
			}
		}

		if selectedNameserver == nil {
			fmt.Printf("⚠️  Selected nameserver (ID: %s) not found.\n", selectedNameserverID)
			fmt.Println("Please select a different nameserver.")
			return nil
		}

		fmt.Printf("Current nameserver: %s (%s)\n", selectedNameserver.DatabaseName, selectedNameserver.ID)
		fmt.Println()
		fmt.Println("You can now use:")
		fmt.Println("  flux-relay sql <query>          # Execute SQL query")
		return nil
	}

	// If argument provided, treat as nameserver selection
	nameserverIdentifier := strings.Join(args, " ")

	// Get all nameservers
	client := api.NewClient(apiURL)
	databasesResponse, err := client.ListDatabases(accessToken, projectID, serverID)
	if err != nil {
		return fmt.Errorf("failed to list nameservers: %w", err)
	}

	// Find nameserver by ID or name (case-insensitive)
	var selectedNameserver *api.Database
	for i := range databasesResponse.Databases {
		ns := &databasesResponse.Databases[i]
		if ns.ID == nameserverIdentifier || 
		   strings.EqualFold(ns.DatabaseName, nameserverIdentifier) {
			selectedNameserver = ns
			break
		}
	}

	if selectedNameserver == nil {
		return fmt.Errorf("nameserver '%s' not found. Use 'flux-relay ns list' to see available nameservers", nameserverIdentifier)
	}

	// Save selected nameserver
	if err := cfg.SetSelectedNameserver(selectedNameserver.ID); err != nil {
		return fmt.Errorf("failed to save nameserver selection: %w", err)
	}

	fmt.Printf("✅ Selected nameserver: %s (%s)\n", selectedNameserver.DatabaseName, selectedNameserver.ID)
	fmt.Println()
	fmt.Println("You can now use:")
	fmt.Println("  flux-relay sql <query>          # Execute SQL query")

	return nil
}

func runNsList(cmd *cobra.Command, args []string) error {
	// Get API URL
	apiURL := getAPIURL()

	// Get access token
	cfg := config.New()
	accessToken := cfg.GetAccessToken()
	if accessToken == "" {
		return fmt.Errorf("not logged in. Run 'flux-relay login' first")
	}

	// Get selected project and server
	projectID := cfg.GetSelectedProject()
	if projectID == "" {
		return fmt.Errorf("no project selected. Use 'flux-relay pr <project-name-or-id>' to select a project")
	}

	serverID := cfg.GetSelectedServer()
	if serverID == "" {
		return fmt.Errorf("no server selected. Use 'flux-relay server <server-name-or-id>' to select a server")
	}

	// Create API client and list nameservers
	client := api.NewClient(apiURL)
	databasesResponse, err := client.ListDatabases(accessToken, projectID, serverID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.Code() == "Unauthorized" || apiErr.Code() == "unauthorized" {
				return fmt.Errorf("authentication failed. Please run 'flux-relay login' again")
			}
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to list nameservers: %w", err)
	}

	nameservers := databasesResponse.Databases

	if len(nameservers) == 0 {
		fmt.Println("No nameservers found in this server.")
		fmt.Println()
		fmt.Println("Create a nameserver using the web dashboard or API.")
		return nil
	}

	// Display nameservers in a table
	fmt.Printf("Found %d nameserver(s) in server:\n\n", len(nameservers))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCREATED\tSTATUS")
	fmt.Fprintln(w, "──\t────\t───────\t──────")

	for _, ns := range nameservers {
		// Format created date
		createdAt, err := time.Parse(time.RFC3339, ns.CreatedAt)
		createdStr := ns.CreatedAt
		if err == nil {
			createdStr = createdAt.Format("2006-01-02")
		}

		// Status
		status := "Active"
		if !ns.IsActive {
			status = "Inactive"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			ns.ID,
			ns.DatabaseName,
			createdStr,
			status,
		)
	}

	w.Flush()
	fmt.Println()

	return nil
}
