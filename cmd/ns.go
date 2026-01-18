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

var nsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new nameserver (database)",
	Long: `Create a new nameserver (database) in the selected server.

The nameserver name must be 1-100 characters and can contain letters, numbers, and underscores.

Examples:
  flux-relay ns create db
  flux-relay ns create my_database`,
	Args: cobra.ExactArgs(1),
	RunE: runNsCreate,
}

var nsInitializeCmd = &cobra.Command{
	Use:   "initialize [nameserver-name-or-id]",
	Short: "Initialize database schema for a nameserver",
	Long: `Initialize the database schema for a nameserver.

This creates the necessary tables for messaging (conversations, messages, users, etc.).
If no nameserver is specified, uses the currently selected nameserver.

Options:
  --type: Schema type - 'messaging' (default), 'analytics', or 'both'
  --drop-existing: Drop existing tables before creating new ones (use with caution!)

Examples:
  flux-relay ns initialize              # Initialize current nameserver
  flux-relay ns initialize db           # Initialize specific nameserver
  flux-relay ns initialize --type both # Initialize with messaging + analytics`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNsInitialize,
}

var schemaType string
var dropExisting bool

func init() {
	nsCmd.AddCommand(nsListCmd)
	nsCmd.AddCommand(nsShellCmd)
	nsCmd.AddCommand(nsCreateCmd)
	nsCmd.AddCommand(nsInitializeCmd)
	
	// Flags for initialize command
	nsInitializeCmd.Flags().StringVar(&schemaType, "type", "messaging", "Schema type: 'messaging', 'analytics', or 'both'")
	nsInitializeCmd.Flags().BoolVar(&dropExisting, "drop-existing", false, "Drop existing tables before creating new ones")
	
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

func runNsCreate(cmd *cobra.Command, args []string) error {
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

	nameserverName := strings.TrimSpace(args[0])
	if nameserverName == "" {
		return fmt.Errorf("nameserver name cannot be empty")
	}

	// Validate name format (1-100 chars, alphanumeric + underscore)
	if len(nameserverName) < 1 || len(nameserverName) > 100 {
		return fmt.Errorf("nameserver name must be 1-100 characters")
	}

	// Create API client and create nameserver
	client := api.NewClient(apiURL)
	fmt.Printf("Creating nameserver '%s'...\n", nameserverName)
	
	response, err := client.CreateNameserver(accessToken, projectID, serverID, nameserverName)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.Code() == "Unauthorized" || apiErr.Code() == "unauthorized" {
				return fmt.Errorf("authentication failed. Please run 'flux-relay login' again")
			}
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to create nameserver: %w", err)
	}

	fmt.Printf("✅ Nameserver created successfully!\n")
	fmt.Printf("   Name: %s\n", response.Database.DatabaseName)
	fmt.Printf("   ID: %s\n", response.Database.ID)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Select this nameserver: flux-relay ns", response.Database.DatabaseName)
	fmt.Println("  2. Initialize schema: flux-relay ns initialize", response.Database.DatabaseName)

	return nil
}

func runNsInitialize(cmd *cobra.Command, args []string) error {
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

	// Determine nameserver ID
	var nameserverID string
	if len(args) > 0 {
		// Nameserver specified as argument
		nameserverIdentifier := strings.Join(args, " ")
		
		// Get all nameservers to find the one specified
		client := api.NewClient(apiURL)
		databasesResponse, err := client.ListDatabases(accessToken, projectID, serverID)
		if err != nil {
			return fmt.Errorf("failed to list nameservers: %w", err)
		}

		// Find nameserver by ID or name (case-insensitive)
		var foundNameserver *api.Database
		for i := range databasesResponse.Databases {
			ns := &databasesResponse.Databases[i]
			if ns.ID == nameserverIdentifier || 
			   strings.EqualFold(ns.DatabaseName, nameserverIdentifier) {
				foundNameserver = ns
				break
			}
		}

		if foundNameserver == nil {
			return fmt.Errorf("nameserver '%s' not found. Use 'flux-relay ns list' to see available nameservers", nameserverIdentifier)
		}
		
		nameserverID = foundNameserver.ID
	} else {
		// Use currently selected nameserver
		nameserverID = cfg.GetSelectedNameserver()
		if nameserverID == "" {
			return fmt.Errorf("no nameserver selected. Use 'flux-relay ns <nameserver-name-or-id>' to select a nameserver, or specify one: flux-relay ns initialize <name>")
		}
	}

	// Validate schema type
	validTypes := map[string]bool{
		"messaging": true,
		"analytics": true,
		"both":      true,
	}
	if !validTypes[schemaType] {
		return fmt.Errorf("invalid schema type '%s'. Must be 'messaging', 'analytics', or 'both'", schemaType)
	}

	// Create API client and initialize nameserver
	client := api.NewClient(apiURL)
	
	// Get nameserver name for display
	databasesResponse, err := client.ListDatabases(accessToken, projectID, serverID)
	if err == nil {
		for _, ns := range databasesResponse.Databases {
			if ns.ID == nameserverID {
				fmt.Printf("Initializing schema for nameserver '%s' (%s)...\n", ns.DatabaseName, nameserverID)
				if dropExisting {
					fmt.Println("⚠️  WARNING: --drop-existing is enabled. Existing tables will be dropped!")
				}
				break
			}
		}
	}

	// Call the API with schema type and drop existing flag
	response, err := client.InitializeNameserverWithOptions(accessToken, projectID, serverID, nameserverID, schemaType, dropExisting)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.Code() == "Unauthorized" || apiErr.Code() == "unauthorized" {
				return fmt.Errorf("authentication failed. Please run 'flux-relay login' again")
			}
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to initialize nameserver: %w", err)
	}

	fmt.Println()
	fmt.Printf("✅ Schema initialized successfully!\n")
	fmt.Printf("   Schema Type: %s\n", response.SchemaType)
	if response.TablesCreated > 0 {
		fmt.Printf("   Tables Created: %d\n", response.TablesCreated)
	}
	if len(response.VerifiedTables) > 0 {
		fmt.Printf("   Verified Tables: %d\n", len(response.VerifiedTables))
	}
	if response.Note != "" {
		fmt.Printf("   Note: %s\n", response.Note)
	}
	fmt.Println()
	fmt.Println("You can now use:")
	fmt.Println("  flux-relay sql \"SELECT * FROM conversations_<nameserver> LIMIT 10\"")

	return nil
}
