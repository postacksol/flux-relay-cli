package cmd

import (
	"fmt"
	"os"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage servers",
	Long:  "List and manage servers in the selected project",
}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all servers in the selected project",
	Long:  "List all servers in the currently selected project with nameserver counts",
	RunE:  runServerList,
}

func init() {
	serverCmd.AddCommand(serverListCmd)
	rootCmd.AddCommand(serverCmd)
}

func runServerList(cmd *cobra.Command, args []string) error {
	// Get API URL
	apiURL := apiBaseURL
	if apiURL == "" {
		apiURL = viper.GetString("api_url")
		if apiURL == "" {
			apiURL = "http://localhost:3000"
		}
	}

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
