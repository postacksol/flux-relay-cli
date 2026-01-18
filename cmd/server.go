package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server [server-name-or-id]",
	Short: "Manage servers",
	Long: `List and select servers in the selected project.

Examples:
  flux-relay server list              # List all servers
  flux-relay server MyServer          # Select by name
  flux-relay server server_123        # Select by ID
  flux-relay server                   # Show current server`,
	Args: cobra.MaximumNArgs(1),
	RunE: runServerShowOrSelect,
}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all servers in the selected project",
	Long:  "List all servers in the currently selected project with nameserver counts",
	RunE:  runServerList,
}

var serverShellCmd = &cobra.Command{
	Use:   "shell [server-name-or-id]",
	Short: "Open interactive SQL shell for a server",
	Long: `Open an interactive SQL shell for a server, similar to Turso's shell.

Examples:
  flux-relay server shell MyServer
  flux-relay srv shell server_123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServerShell(args[0])
	},
}

func init() {
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverShellCmd)
	rootCmd.AddCommand(serverCmd)
	
	// Add 'srv' as an alias for 'server'
	srvCmd := *serverCmd
	srvCmd.Use = "srv [server-name-or-id]"
	rootCmd.AddCommand(&srvCmd)
}

func runServerList(cmd *cobra.Command, args []string) error {
	// Get API URL
	apiURL := getAPIURL()

	// Get access token
	cfg := config.New()
	accessToken := cfg.GetAccessToken()
	if accessToken == "" {
		return fmt.Errorf("not logged in. Run 'flux-relay login' first")
	}

	// Get selected project
	projectID := cfg.GetSelectedProject()
	if projectID == "" {
		return fmt.Errorf("no project selected. Use 'flux-relay pr <project-name-or-id>' to select a project")
	}

	// Create API client and list servers
	client := api.NewClient(apiURL)
	serversResponse, err := client.ListServers(accessToken, projectID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.Code() == "Unauthorized" || apiErr.Code() == "unauthorized" {
				return fmt.Errorf("authentication failed. Please run 'flux-relay login' again")
			}
			if apiErr.Code() == "Project not found" || apiErr.Code() == "project not found" {
				return fmt.Errorf("project not found. Use 'flux-relay pr <project-name-or-id>' to select a valid project")
			}
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to list servers: %w", err)
	}

	servers := serversResponse.Servers

	if len(servers) == 0 {
		fmt.Println("No servers found in this project.")
		fmt.Println()
		fmt.Println("Create a server using the web dashboard or API.")
		return nil
	}

	// Get nameserver counts for each server in parallel
	type serverWithCount struct {
		Server        api.Server
		NameserverCount int
	}

	var wg sync.WaitGroup
	serversWithCounts := make([]serverWithCount, len(servers))
	errors := make([]error, len(servers))

	for i, server := range servers {
		wg.Add(1)
		go func(idx int, srv api.Server) {
			defer wg.Done()
			databasesResponse, err := client.ListDatabases(accessToken, projectID, srv.ID)
			if err != nil {
				errors[idx] = err
				serversWithCounts[idx] = serverWithCount{
					Server:          srv,
					NameserverCount: 0, // Default to 0 on error
				}
				return
			}
			// Count active databases (nameservers)
			count := 0
			for _, db := range databasesResponse.Databases {
				if db.IsActive {
					count++
				}
			}
			serversWithCounts[idx] = serverWithCount{
				Server:          srv,
				NameserverCount: count,
			}
		}(i, server)
	}

	wg.Wait()

	// Display servers in a table
	fmt.Printf("Found %d server(s) in project:\n\n", len(servers))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION\tNAMESERVERS\tCREATED\tSTATUS")
	fmt.Fprintln(w, "──\t────\t───────────\t───────────\t───────\t──────")

	for _, item := range serversWithCounts {
		server := item.Server

		// Format created date
		createdAt, err := time.Parse(time.RFC3339, server.CreatedAt)
		createdStr := server.CreatedAt
		if err == nil {
			createdStr = createdAt.Format("2006-01-02")
		}

		// Truncate description if too long
		description := server.Description
		if len(description) > 30 {
			description = description[:27] + "..."
		}
		if description == "" {
			description = "-"
		}

		// Status
		status := "Active"
		if !server.IsActive {
			status = "Inactive"
		}

		// Nameserver count
		nameserverCount := fmt.Sprintf("%d", item.NameserverCount)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			server.ID,
			server.Name,
			description,
			nameserverCount,
			createdStr,
			status,
		)
	}

	w.Flush()
	fmt.Println()

	// Show any errors (non-critical, just warn)
	hasErrors := false
	for i, err := range errors {
		if err != nil {
			if !hasErrors {
				fmt.Println("⚠️  Warnings:")
				hasErrors = true
			}
			fmt.Printf("  Could not get nameserver count for server %s: %v\n", servers[i].ID, err)
		}
	}
	if hasErrors {
		fmt.Println()
	}

	return nil
}

func runServerShowOrSelect(cmd *cobra.Command, args []string) error {
	// Get API URL
	apiURL := getAPIURL()

	// Get access token
	cfg := config.New()
	accessToken := cfg.GetAccessToken()
	if accessToken == "" {
		return fmt.Errorf("not logged in. Run 'flux-relay login' first")
	}

	// Get selected project
	projectID := cfg.GetSelectedProject()
	if projectID == "" {
		return fmt.Errorf("no project selected. Use 'flux-relay pr <project-name-or-id>' to select a project")
	}

	// If no argument, show current server
	if len(args) == 0 {
		selectedServerID := cfg.GetSelectedServer()
		if selectedServerID == "" {
			fmt.Println("No server selected.")
			fmt.Println()
			fmt.Println("Select a server using:")
			fmt.Println("  flux-relay server <server-name-or-id>")
			fmt.Println("  flux-relay srv <server-name-or-id>")
			fmt.Println()
			fmt.Println("Or list available servers:")
			fmt.Println("  flux-relay server list")
			return nil
		}

		// Get server details
		client := api.NewClient(apiURL)
		serversResponse, err := client.ListServers(accessToken, projectID)
		if err != nil {
			return fmt.Errorf("failed to get server info: %w", err)
		}

		// Find the selected server
		var selectedServer *api.Server
		for i := range serversResponse.Servers {
			if serversResponse.Servers[i].ID == selectedServerID {
				selectedServer = &serversResponse.Servers[i]
				break
			}
		}

		if selectedServer == nil {
			fmt.Printf("⚠️  Selected server (ID: %s) not found.\n", selectedServerID)
			fmt.Println("Please select a different server.")
			return nil
		}

		fmt.Printf("Current server: %s (%s)\n", selectedServer.Name, selectedServer.ID)
		if selectedServer.Description != "" {
			fmt.Printf("Description: %s\n", selectedServer.Description)
		}
		
		// Show selected nameserver if any
		selectedNameserverID := cfg.GetSelectedNameserver()
		if selectedNameserverID != "" {
			databasesResponse, err := client.ListDatabases(accessToken, projectID, selectedServerID)
			if err == nil {
				for _, db := range databasesResponse.Databases {
					if db.ID == selectedNameserverID {
						fmt.Printf("Nameserver: %s (%s)\n", db.DatabaseName, db.ID)
						break
					}
				}
			}
		}
		return nil
	}

	// If argument provided, treat as server selection
	serverIdentifier := strings.Join(args, " ")

	// Get all servers
	client := api.NewClient(apiURL)
	serversResponse, err := client.ListServers(accessToken, projectID)
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}

	// Find server by ID or name (case-insensitive)
	var selectedServer *api.Server
	for i := range serversResponse.Servers {
		server := &serversResponse.Servers[i]
		if server.ID == serverIdentifier || 
		   strings.EqualFold(server.Name, serverIdentifier) {
			selectedServer = server
			break
		}
	}

	if selectedServer == nil {
		return fmt.Errorf("server '%s' not found. Use 'flux-relay server list' to see available servers", serverIdentifier)
	}

	// Save selected server
	if err := cfg.SetSelectedServer(selectedServer.ID); err != nil {
		return fmt.Errorf("failed to save server selection: %w", err)
	}

	fmt.Printf("✅ Selected server: %s (%s)\n", selectedServer.Name, selectedServer.ID)
	if selectedServer.Description != "" {
		fmt.Printf("   Description: %s\n", selectedServer.Description)
	}
	fmt.Println()
	fmt.Println("You can now use:")
	fmt.Println("  flux-relay ns list              # List nameservers")
	fmt.Println("  flux-relay ns <nameserver-name> # Select nameserver")
	fmt.Println("  flux-relay sql <query>          # Execute SQL query")

	return nil
}
